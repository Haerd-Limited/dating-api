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
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=auth
type Service interface {
	VerifyCode(ctx context.Context, in domain.VerifyCode) (*domain.AuthResult, error)
	RequestCode(ctx context.Context, requestCodeDetails domain.RequestCode) (*domain.RequestCodeResult, error)
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
	env                  string
}

func NewAuthService(
	logger *zap.Logger,
	jwtSecret string,
	UserService user.Service,
	AuthRepository authStorage.AuthRepository,
	awsService aws.Service,
	communicationService communication.Service,
	env string,
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
		env:                  env,
	}
}

var (
	ErrRefreshTokenExpired        = errors.New("refresh token expired")
	ErrRefreshTokenRevoked        = errors.New("refresh token has been revoked")
	ErrRefreshTokenAlreadyRevoked = errors.New("refresh token already revoked")
	ErrPhoneNumberRequired        = errors.New("phone number required")
	ErrUserAlreadyRegistered      = errors.New("user already registered")
	ErrEmailRequired              = errors.New("email required")
	ErrInvalidChannel             = errors.New("invalid channel")
)

func (as *authService) VerifyCode(ctx context.Context, in domain.VerifyCode) (*domain.AuthResult, error) {
	if in.Purpose == "" {
		in.Purpose = constants.LoginPurpose
	}

	identifier, err := as.normalizeIdentifier(in.Channel, in.Email, in.Phone)
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "invalid identifier", err, zap.String("channel", in.Channel))
	}

	purpose := strings.ToLower(in.Purpose)

	// Reduce twilio usage when testing
	if strings.ToLower(as.env) == constants.ProductionEnvironment {
		// 1) Find latest active code
		rec, err := as.AuthRepo.FindActiveVerificationCode(ctx, in.Channel, identifier, purpose)
		if err != nil {
			// do not reveal which part failed
			return nil, commonlogger.LogError(as.logger, "failed to find active code", err, zap.String("channel", in.Channel), zap.String("purpose", purpose))
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
	}

	// 4) Resolve user (create on register )
	exists, err := as.UserService.UserExistsByIdentifier(ctx, in.Channel, identifier)
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "user lookup failed", err, zap.String("channel", in.Channel))
	}

	var userDetails *userdomain.User

	switch purpose {
	case constants.LoginPurpose:
		if !exists {
			return nil, errors.New("invalid or expired code")
		}

		if in.Channel == constants.SmsChannel {
			userDetails, err = as.UserService.GetUserByPhoneNumber(ctx, identifier)
			if err != nil {
				return nil, commonlogger.LogError(as.logger, "auth user", err, zap.String("channel", in.Channel))
			}
		} else {
			return nil, ErrInvalidChannel
		}

		tokens, tokenErr := as.GenerateAccessAndRefreshToken(ctx, userDetails.ID)
		if tokenErr != nil {
			return nil, commonlogger.LogError(as.logger, "issue tokens", tokenErr, zap.String("userID", userDetails.ID))
		}

		return &domain.AuthResult{
			RefreshToken: tokens.RefreshToken,
			AccessToken:  tokens.AccessToken,
			User: &domain.User{
				ID:             userDetails.ID,
				OnboardingStep: onboardingdomain.Steps(userDetails.OnboardingStep),
			},
		}, nil
	case constants.RegisterPurpose:
		if exists {
			// User exists - check if they're still in pre-registration
			if in.Channel == constants.SmsChannel {
				userDetails, getUserErr := as.UserService.GetUserByPhoneNumber(ctx, identifier)
				if getUserErr != nil {
					return nil, commonlogger.LogError(as.logger, "failed to get user", getUserErr, zap.String("channel", in.Channel))
				}

				currentStep := onboardingdomain.Steps(userDetails.OnboardingStep)
				// If user is still in pre-registration (INTRO or BASICS), allow them to continue
				if currentStep == onboardingdomain.OnboardingStepsIntro || currentStep == onboardingdomain.OnboardingStepsBasics {
					tokens, tokenErr := as.GenerateAccessAndRefreshToken(ctx, userDetails.ID)
					if tokenErr != nil {
						return nil, commonlogger.LogError(as.logger, "failed to generate tokens", tokenErr, zap.String("userID", userDetails.ID))
					}

					return &domain.AuthResult{
						RefreshToken: tokens.RefreshToken,
						AccessToken:  tokens.AccessToken,
						User: &domain.User{
							ID:             userDetails.ID,
							OnboardingStep: currentStep,
						},
					}, nil
				}
				// User has completed pre-registration, they should use login instead
				return nil, ErrUserAlreadyRegistered
			} else {
				return nil, ErrInvalidChannel
			}
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
			return nil, commonlogger.LogError(as.logger, "failed to create user", regErr)
		}

		tokens, tokenErr := as.GenerateAccessAndRefreshToken(ctx, userID)
		if tokenErr != nil {
			return nil, commonlogger.LogError(as.logger, "failed to generate tokens", tokenErr, zap.String("userID", userID))
		}

		return &domain.AuthResult{
			RefreshToken: tokens.RefreshToken,
			AccessToken:  tokens.AccessToken,
			User: &domain.User{
				ID:             userID,
				OnboardingStep: onboardingdomain.OnboardingStepsIntro,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown purpose: %s", purpose)
	}
}

func (as *authService) RequestCode(ctx context.Context, requestCodeDetails domain.RequestCode) (*domain.RequestCodeResult, error) {
	if requestCodeDetails.Purpose == "" {
		requestCodeDetails.Purpose = constants.LoginPurpose
	}

	identifier, err := as.normalizeIdentifier(requestCodeDetails.Channel, requestCodeDetails.Email, requestCodeDetails.Phone)
	if err != nil {
		return &domain.RequestCodeResult{SentTo: ""}, commonlogger.LogError(as.logger, "failed to normalize identifier", err, zap.String("channel", requestCodeDetails.Channel))
	}

	mask := maskIdentifier(requestCodeDetails.Channel, identifier)
	purpose := strings.ToLower(requestCodeDetails.Purpose)

	// Rate limits (per identifier + per IP), but do not leak to client
	window := time.Now().Add(-1 * time.Hour).UTC()

	n1, _ := as.AuthRepo.CountRecentSends(ctx, requestCodeDetails.Channel, identifier, purpose, window)
	if n1 >= as.perIDPerHour {
		// pretend success
		return &domain.RequestCodeResult{SentTo: mask}, fmt.Errorf("rate limit exceeded")
	}

	n2, _ := as.AuthRepo.CountRecentSendsByIP(ctx, clientIP(requestCodeDetails.IP), window)
	if n2 >= as.perIPPerHour {
		return &domain.RequestCodeResult{SentTo: mask}, fmt.Errorf("rate limit exceeded")
	}

	exists, err := as.UserService.UserExistsByIdentifier(ctx, requestCodeDetails.Channel, identifier)
	if err != nil {
		return &domain.RequestCodeResult{SentTo: mask}, commonlogger.LogError(as.logger, "existence check failed", err, zap.String("channel", requestCodeDetails.Channel))
	}

	var onboardingStep *onboardingdomain.Steps

	// For register purpose, if user exists, check if they're still in pre-registration
	if purpose == constants.RegisterPurpose && exists {
		if requestCodeDetails.Channel == constants.SmsChannel {
			userDetails, getUserErr := as.UserService.GetUserByPhoneNumber(ctx, identifier)
			if getUserErr == nil && userDetails != nil {
				step := onboardingdomain.Steps(userDetails.OnboardingStep)
				// Only return step if user is still in pre-registration phase (INTRO or BASICS)
				if step == onboardingdomain.OnboardingStepsIntro || step == onboardingdomain.OnboardingStepsBasics {
					onboardingStep = &step
					// Allow sending code so they can continue pre-registration
				}
			}
		}
	}

	shouldSend := false

	switch purpose {
	case constants.LoginPurpose:
		shouldSend = exists
	case constants.RegisterPurpose:
		// Send code if user doesn't exist, or if user exists and is in pre-registration
		shouldSend = !exists || (exists && onboardingStep != nil)
	default:
		// unknown purpose -> treat as login
		shouldSend = exists
	}

	if !shouldSend {
		// Pretend success, do nothing further
		return &domain.RequestCodeResult{SentTo: mask, OnboardingStep: onboardingStep}, fmt.Errorf("not sending code")
	}

	// Generate a 6-digit numeric code
	code, err := RandomDigits(6)
	if err != nil {
		return &domain.RequestCodeResult{SentTo: mask, OnboardingStep: onboardingStep}, commonlogger.LogError(as.logger, "failed to generate code", err, zap.String("channel", requestCodeDetails.Channel))
	}

	// Hash the code (never store plaintext)
	hash := as.hmac(code, identifier, purpose)

	rec := mapper.ToVerificationCodeEntity(&domain.VerificationCode{
		Channel:     requestCodeDetails.Channel,
		Identifier:  identifier,
		Purpose:     purpose,
		CodeHash:    hash,
		ExpiresAt:   time.Now().Add(as.codeTTL).UTC(),
		RequestIP:   net.ParseIP(clientIP(requestCodeDetails.IP)),
		MaxAttempts: 5,
	})

	err = as.AuthRepo.InsertVerificationCode(ctx, rec)
	if err != nil {
		// still mask success to caller
		return &domain.RequestCodeResult{SentTo: mask, OnboardingStep: onboardingStep}, commonlogger.LogError(as.logger, "failed to insert verification code", err, zap.String("channel", requestCodeDetails.Channel))
	}

	// Deliver
	var sendErr error

	switch requestCodeDetails.Channel {
	case constants.EmailChannel:
		sendErr = as.communicationService.SendEmailOTP(identifier, code)
	case constants.SmsChannel:
		sendErr = as.communicationService.SendSMSOTP(identifier, code)
	}

	if sendErr != nil {
		return &domain.RequestCodeResult{SentTo: mask, OnboardingStep: onboardingStep}, commonlogger.LogError(as.logger, "failed to send verification code", sendErr, zap.String("channel", requestCodeDetails.Channel))
	}

	return &domain.RequestCodeResult{
		SentTo:         mask,
		OnboardingStep: onboardingStep,
	}, nil
}

func (as *authService) RefreshToken(ctx context.Context, refreshInput domain.Refresh) (*domain.AuthResult, error) {
	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, refreshInput.RefreshToken)
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to get refresh token", err)
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	if refreshToken.Revoked {
		return nil, ErrRefreshTokenRevoked
	}

	err = as.AuthRepo.RevokeRefreshToken(ctx, refreshToken.ID)
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to revoke token", err, zap.String("tokenID", refreshToken.ID))
	}

	accessToken, err := auth.GenerateAccessToken(refreshToken.UserID, []byte(as.jwtSecret))
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to generate access token", err, zap.String("userID", refreshToken.UserID))
	}

	newRefreshToken := auth.GenerateRefreshToken(refreshToken.UserID)

	newRefreshTokenEntity := mapper.ToRefreshTokenEntity(newRefreshToken)

	err = as.AuthRepo.InsertRefreshToken(ctx, newRefreshTokenEntity)
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to store new refresh token", err, zap.String("userID", refreshToken.UserID))
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
		return commonlogger.LogError(as.logger, "failed to revoke token", err, zap.String("tokenID", refreshToken.ID))
	}

	return nil
}

func (as *authService) GenerateAccessAndRefreshToken(ctx context.Context, userID string) (*domain.AuthResult, error) {
	accessToken, err := auth.GenerateAccessToken(userID, []byte(as.jwtSecret))
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to generate access token", err, zap.String("userID", userID))
	}

	refreshToken := auth.GenerateRefreshToken(userID)

	err = as.AuthRepo.InsertRefreshToken(ctx, mapper.ToRefreshTokenEntity(refreshToken))
	if err != nil {
		return nil, commonlogger.LogError(as.logger, "failed to insert refresh token", err, zap.String("userID", refreshToken.UserID))
	}

	return &domain.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (as *authService) normalizeIdentifier(channel string, email *string, phone *string) (string, error) {
	switch channel {
	case constants.EmailChannel:
		if email == nil || *email == "" {
			return "", errors.New("email required")
		}

		e := strings.TrimSpace(strings.ToLower(*email))
		// cheap sanity
		if !strings.Contains(e, "@") {
			return "", errors.New("invalid email")
		}

		return e, nil
	case constants.SmsChannel:
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
		return "", ErrInvalidChannel
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
