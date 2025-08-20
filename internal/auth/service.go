package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
	"time"

	"go.uber.org/zap"

	authDomain "github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/internal/auth/mapper"
	authStorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/auth"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=auth
type Service interface {
	Register(ctx context.Context, registerDetails *authDomain.Register) (*authDomain.AuthTokensAndUserID, error)
	Login(ctx context.Context, loginInput authDomain.Login) (*authDomain.AuthTokensAndUserID, error)
	RefreshToken(ctx context.Context, refreshInput authDomain.Refresh) (*authDomain.AuthTokensAndUserID, error)
	RevokeRefreshToken(ctx context.Context, revokeRefreshTokenInput authDomain.RevokeRefreshToken) error
}

type authService struct {
	logger      *zap.Logger
	jwtSecret   string
	UserService user.Service
	AuthRepo    authStorage.AuthRepository
	awsService  aws.AWSService
}

func NewAuthService(
	logger *zap.Logger,
	jwtSecret string,
	UserService user.Service,
	AuthRepository authStorage.AuthRepository,
	awsService aws.AWSService,
) Service {
	return &authService{
		logger:      logger,
		jwtSecret:   jwtSecret,
		UserService: UserService,
		AuthRepo:    AuthRepository,
		awsService:  awsService,
	}
}

var (
	ErrRefreshTokenExpired        = errors.New("refresh token expired")
	ErrRefreshTokenRevoked        = errors.New("refresh token has been revoked")
	ErrRefreshTokenAlreadyRevoked = errors.New("refresh token already revoked")
)

func (as *authService) Register(ctx context.Context, registerDetails *authDomain.Register) (*authDomain.AuthTokensAndUserID, error) {
	createUserResult, err := as.UserService.CreateUser(ctx, &domain.User{
		Email:       registerDetails.Email,
		PhoneNumber: registerDetails.PhoneNumber,
		FirstName:   registerDetails.FirstName,
		LastName:    registerDetails.LastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := auth.GenerateAccessToken(createUserResult.UserID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token userID=%s: %w", createUserResult.UserID, err)
	}

	refreshToken := auth.GenerateRefreshToken(createUserResult.UserID)

	err = as.AuthRepo.InsertRefreshToken(ctx, mapper.ToRefreshTokenEntity(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token userID=%s: %w", refreshToken.UserID, err)
	}

	return &authDomain.AuthTokensAndUserID{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		UserID:       createUserResult.UserID,
	}, nil
}

func (as *authService) Login(ctx context.Context, loginDetails authDomain.Login) (*authDomain.AuthTokensAndUserID, error) {
	userDetails, err := as.UserService.AuthenticateUser(ctx, utils.Redacted(loginDetails.PhoneNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user phone=%s: %w", loginDetails.PhoneNumber, err)
	}

	// Revoke/delete other user associated refresh tokens
	err = as.AuthRepo.RevokeAllRefreshTokens(ctx, userDetails.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke all refresh tokens userID=%s: %w", userDetails.ID, err)
	}

	accessToken, err := auth.GenerateAccessToken(userDetails.ID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token userID=%s: %w", userDetails.ID, err)
	}

	refreshToken := auth.GenerateRefreshToken(userDetails.ID)

	refreshTokenEntity := mapper.ToRefreshTokenEntity(refreshToken)

	err = as.AuthRepo.InsertRefreshToken(ctx, refreshTokenEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token userID=%s token=%s: %w", refreshToken.UserID, utils.Redacted(refreshToken.Token), err)
	}

	return &authDomain.AuthTokensAndUserID{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		UserID:       userDetails.ID,
	}, nil
}

func (as *authService) RefreshToken(ctx context.Context, refreshInput authDomain.Refresh) (*authDomain.AuthTokensAndUserID, error) {
	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, refreshInput.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token refreshToken=%s: %w", utils.Redacted(refreshInput.RefreshToken), err)
	}

	if time.Now().After(refreshToken.ExpiresAt) { //todo: move to repository layer
		return nil, ErrRefreshTokenExpired
	}

	if refreshToken.Revoked { //todo: move to repository layer
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

	return &authDomain.AuthTokensAndUserID{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
	}, nil
}

func (as *authService) RevokeRefreshToken(ctx context.Context, revokeRefreshTokenInput authDomain.RevokeRefreshToken) error {
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
