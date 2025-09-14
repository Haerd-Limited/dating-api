package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/internal/auth/mapper"
	authStorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/communication"
	onboardingdomain "github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/auth"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=auth
type Service interface {
	VerifyCode(ctx context.Context, in domain.VerifyCode) (*domain.AuthResult, error)
	RequestCode(ctx context.Context, requestCodeDetails domain.RequestCode) (string, error)
	RefreshToken(ctx context.Context, refreshInput domain.Refresh) (*domain.AuthResult, error)
	RevokeRefreshToken(ctx context.Context, revokeRefreshTokenInput domain.RevokeRefreshToken) error
	GenerateAccessAndRefreshToken(ctx context.Context, userID string) (*domain.AuthResult, error)
}

type authService struct {
	logger               *zap.Logger
	jwtSecret            string
	UserService          user.Service
	AuthRepo             authStorage.AuthRepository
	awsService           aws.Service
	communicationService communication.Service
	codeTTL              time.Duration
	perIDPerHour         int // e.g., 3
	perIPPerHour         int // e.g., 20
}

func NewAuthService(
	logger *zap.Logger,
	jwtSecret string,
	UserService user.Service,
	AuthRepository authStorage.AuthRepository,
	awsService aws.Service,
	communicationService communication.Service,
) Service {
	return &authService{
		logger:               logger,
		jwtSecret:            jwtSecret,
		UserService:          UserService,
		AuthRepo:             AuthRepository,
		awsService:           awsService,
		communicationService: communicationService,
		codeTTL:              10 * time.Minute,
		perIDPerHour:         3,
		perIPPerHour:         20,
	}
}

var (
	ErrRefreshTokenExpired        = errors.New("refresh token expired")
	ErrRefreshTokenRevoked        = errors.New("refresh token has been revoked")
	ErrRefreshTokenAlreadyRevoked = errors.New("refresh token already revoked")
	ErrPhoneNumberRequired        = errors.New("phone number required")
	ErrEmailRequired              = errors.New("email required")
)

func (as *authService) VerifyCode(ctx context.Context, in domain.VerifyCode) (*domain.AuthResult, error) {
	identifier, err := as.normalizeIdentifier(in.Channel, in.Email, in.Phone)
	if err != nil {
		return nil, fmt.Errorf("invalid identifier: %w", err)
	}

	purpose := strings.ToLower(in.Purpose)

	// 1) Find latest active code
	rec, err := as.AuthRepo.FindActiveVerificationCode(ctx, in.Channel, identifier, purpose)
	if err != nil {
		// do not reveal which part failed
		return nil, fmt.Errorf("failed to find active code channel=%s identifier=%s purpose=%s: %w",
			in.Channel, identifier, purpose, err)
	}

	if rec.Attempts >= rec.MaxAttempts {
		return nil, errors.New("too many attempts")
	}

	// 2) Compare HMAC
	expected := as.hmac(in.Code, identifier, purpose)
	if !hmac.Equal([]byte(expected), []byte(rec.CodeHash)) {
		_ = as.AuthRepo.IncrementAttempts(ctx, rec.ID)
		return nil, errors.New("invalid or expired code")
	}

	// 3) Consume single-use
	ok, err := as.AuthRepo.ConsumeVerificationCode(ctx, rec.ID)
	if err != nil || !ok {
		return nil, errors.New("invalid or expired code")
	}

	// 4) Resolve user (create on register )
	exists, err := as.UserService.UserExistsByIdentifier(ctx, in.Channel, identifier)
	if err != nil {
		return nil, fmt.Errorf("user lookup failed: %w", err)
	}

	var userDetails *userdomain.User

	switch purpose {
	case "login":
		if !exists {
			return nil, errors.New("invalid or expired code")
		}

		userDetails, err = as.UserService.AuthenticateUser(ctx, identifier)
		if err != nil {
			return nil, fmt.Errorf("auth user: %w", err)
		}

		toks, err := as.GenerateAccessAndRefreshToken(ctx, userDetails.ID)
		if err != nil {
			return nil, fmt.Errorf("issue tokens: %w", err)
		}

		return &domain.AuthResult{
			RefreshToken: toks.RefreshToken,
			AccessToken:  toks.AccessToken,
			User: &domain.User{
				ID:             userDetails.ID,
				OnboardingStep: onboardingdomain.Steps(userDetails.OnboardingStep),
			},
		}, nil
	case "register":
		if exists {
			return nil, errors.New("user already exists")
		}

		if in.Phone == nil {
			return nil, ErrPhoneNumberRequired
		}

		// create user
		userID, regErr := as.UserService.CreateUser(ctx, userdomain.User{
			PhoneNumber:    *in.Phone,
			OnboardingStep: string(onboardingdomain.OnboardingStepsIntro),
		})
		if regErr != nil {
			return nil, fmt.Errorf("failed to create user: %w", regErr)
		}

		toks, regErr := as.GenerateAccessAndRefreshToken(ctx, *userID)
		if regErr != nil {
			return nil, fmt.Errorf("failed to generate tokens: %w", err)
		}

		return &domain.AuthResult{
			RefreshToken: toks.RefreshToken,
			AccessToken:  toks.AccessToken,
			User: &domain.User{
				ID:             *userID,
				OnboardingStep: onboardingdomain.OnboardingStepsIntro,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown purpose: %s", purpose)
	}
}

func (as *authService) RequestCode(ctx context.Context, requestCodeDetails domain.RequestCode) (string, error) {
	identifier, err := as.normalizeIdentifier(requestCodeDetails.Channel, requestCodeDetails.Email, requestCodeDetails.Phone)
	if err != nil {
		return "", fmt.Errorf("failed to normalize identifier: %w", err)
	}

	mask := maskIdentifier(requestCodeDetails.Channel, identifier)
	purpose := strings.ToLower(requestCodeDetails.Purpose)

	// Rate limits (per identifier + per IP), but do not leak to client
	window := time.Now().Add(-1 * time.Hour)

	n1, _ := as.AuthRepo.CountRecentSends(ctx, requestCodeDetails.Channel, identifier, purpose, window)
	if n1 >= as.perIDPerHour {
		// pretend success
		return mask, fmt.Errorf("rate limit exceeded")
	}

	n2, _ := as.AuthRepo.CountRecentSendsByIP(ctx, clientIP(requestCodeDetails.IP), window)
	if n2 >= as.perIPPerHour {
		return mask, fmt.Errorf("rate limit exceeded")
	}

	exists, err := as.UserService.UserExistsByIdentifier(ctx, requestCodeDetails.Channel, identifier)
	if err != nil {
		return mask, fmt.Errorf("existence check failed channel=%s identifier=%s : %w",
			requestCodeDetails.Channel, identifier, err)
	}

	shouldSend := false

	switch purpose {
	case "login":
		shouldSend = exists
	case "register":
		shouldSend = !exists
	default:
		// unknown purpose -> treat as login
		shouldSend = exists
	}

	if !shouldSend {
		// Pretend success, do nothing further
		return mask, fmt.Errorf("not sending code")
	}

	// Generate a 6-digit numeric code
	code, err := RandomDigits(6)
	if err != nil {
		return mask, fmt.Errorf("failed to generate code: %w", err)
	}

	// Hash the code (never store plaintext)
	hash := as.hmac(code, identifier, purpose)

	rec := mapper.ToVerificationCodeEntity(&domain.VerificationCode{
		Channel:     requestCodeDetails.Channel,
		Identifier:  identifier,
		Purpose:     purpose,
		CodeHash:    hash,
		ExpiresAt:   time.Now().Add(as.codeTTL),
		RequestIP:   net.ParseIP(clientIP(requestCodeDetails.IP)),
		MaxAttempts: 5,
	})

	err = as.AuthRepo.InsertVerificationCode(ctx, rec)
	if err != nil {
		// still mask success to caller
		return mask, fmt.Errorf("failed to insert verification code: %w", err)
	}

	// Deliver
	var sendErr error

	switch requestCodeDetails.Channel {
	case "email":
		sendErr = as.communicationService.SendEmailOTP(identifier, code)
	case "sms":
		sendErr = as.communicationService.SendSMSOTP(identifier, code)
	}

	if sendErr != nil {
		return mask, fmt.Errorf("failed to send verification code channel=%s identifier=%s: %w",
			requestCodeDetails.Channel, identifier, sendErr)
	}

	return mask, nil
}

func (as *authService) RefreshToken(ctx context.Context, refreshInput domain.Refresh) (*domain.AuthResult, error) {
	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, refreshInput.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token refreshToken=%s: %w", utils.Redacted(refreshInput.RefreshToken), err)
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	if refreshToken.Revoked {
		return nil, ErrRefreshTokenRevoked
	}

	err = as.AuthRepo.RevokeRefreshToken(ctx, refreshToken.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	accessToken, err := auth.GenerateAccessToken(refreshToken.UserID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token userID=%s: %w", refreshToken.UserID, err)
	}

	newRefreshToken := auth.GenerateRefreshToken(refreshToken.UserID)

	newRefreshTokenEntity := mapper.ToRefreshTokenEntity(newRefreshToken)

	err = as.AuthRepo.InsertRefreshToken(ctx, newRefreshTokenEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token userID=%s: %w", refreshToken.UserID, err)
	}

	return &domain.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
	}, nil
}

func (as *authService) RevokeRefreshToken(ctx context.Context, revokeRefreshTokenInput domain.RevokeRefreshToken) error {
	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, revokeRefreshTokenInput.RefreshToken)
	if err != nil {
		return err
	}

	if refreshToken.Revoked {
		return ErrRefreshTokenAlreadyRevoked
	}

	err = as.AuthRepo.RevokeRefreshToken(ctx, refreshToken.ID)
	if err != nil {
		return fmt.Errorf("failed to revoke token id=%s: %w", refreshToken.ID, err)
	}

	return nil
}

func (as *authService) GenerateAccessAndRefreshToken(ctx context.Context, userID string) (*domain.AuthResult, error) {
	accessToken, err := auth.GenerateAccessToken(userID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token userID=%s: %w", userID, err)
	}

	refreshToken := auth.GenerateRefreshToken(userID)

	err = as.AuthRepo.InsertRefreshToken(ctx, mapper.ToRefreshTokenEntity(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token userID=%s: %w", refreshToken.UserID, err)
	}

	return &domain.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (as *authService) normalizeIdentifier(channel string, email *string, phone *string) (string, error) {
	switch channel {
	case "email":
		if email == nil || *email == "" {
			return "", errors.New("email required")
		}

		e := strings.TrimSpace(strings.ToLower(*email))
		// cheap sanity
		if !strings.Contains(e, "@") {
			return "", errors.New("invalid email")
		}

		return e, nil
	case "sms":
		if phone == nil || *phone == "" {
			return "", errors.New("phone required")
		}

		p := strings.TrimSpace(*phone)
		// quick E.164 check
		ok, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, p)
		if !ok {
			return "", errors.New("invalid phone (E.164)")
		}

		return p, nil
	default:
		return "", errors.New("invalid channel")
	}
}

func (as *authService) hmac(code, identifier, purpose string) string {
	m := hmac.New(sha256.New, []byte(as.jwtSecret)) // todo: replace with signing secret and add to env file
	m.Write([]byte(identifier))
	m.Write([]byte{0})
	m.Write([]byte(purpose))
	m.Write([]byte{0})
	m.Write([]byte(code))

	return hex.EncodeToString(m.Sum(nil))
}

func maskIdentifier(channel, id string) string {
	if channel == "email" {
		parts := strings.Split(id, "@")
		if len(parts) != 2 {
			return "***"
		}

		local := parts[0]
		if len(local) > 1 {
			local = local[:1] + strings.Repeat("*", len(local)-1)
		} else {
			local = "*"
		}

		return local + "@" + parts[1]
	}
	// phone: keep country code and last 2
	if len(id) > 4 {
		return id[:3] + strings.Repeat("*", len(id)-5) + id[len(id)-2:]
	}

	return "***"
}

func clientIP(raw string) string {
	// crude parse of "ip:port" or CSV XFF; good enough for rate limiting
	if idx := strings.IndexByte(raw, ','); idx > 0 {
		raw = raw[:idx]
	}

	if idx := strings.LastIndexByte(raw, ':'); idx > 0 {
		return raw[:idx]
	}

	return raw
}

// RandomDigits generates a string with n random digits (0–9).
func RandomDigits(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("n must be > 0")
	}

	const digits = "0123456789"

	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random digits: %w", err)
	}

	for i := 0; i < n; i++ {
		b[i] = digits[int(b[i])%10]
	}

	return string(b), nil
}
