package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	authDomain "github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/internal/auth/mapper"
	authStorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	awsDomain "github.com/Haerd-Limited/dating-api/internal/aws/domain"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/auth"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=auth
type AuthService interface {
	Register(ctx context.Context, registerInput *authDomain.Register) (*authDomain.AuthTokensAndUser, error)
	Login(ctx context.Context, loginInput authDomain.Login) (*authDomain.AuthTokensAndUser, error)
	RefreshToken(ctx context.Context, refreshInput authDomain.Refresh) (*authDomain.AuthTokens, error)
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
) AuthService {
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
	ErrRefreshTokenNotFound       = errors.New("refresh token not found")
)

func (as *authService) Register(ctx context.Context, registerInput *authDomain.Register) (*authDomain.AuthTokensAndUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerInput.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userDomain := &domain.User{
		FullName:       registerInput.FullName,
		Username:       registerInput.Username,
		HashedPassword: string(hashedPassword),
		Email:          registerInput.Email,
		Dob:            registerInput.DateOfBirth,
		Bio:            registerInput.Bio,
		Gender:         registerInput.Gender,
	}

	if registerInput.ProfileImage != nil && registerInput.ProfileImageHeader != nil {
		// upload post image to s3
		userDomain.ProfileImageURL, err = as.awsService.UploadImage(ctx, awsDomain.ImageUpload{
			AuthorID:    registerInput.Username,
			ImageHeader: *registerInput.ProfileImageHeader,
			ImageFile:   *registerInput.ProfileImage,
			FolderPath:  awsDomain.FolderProfilePictures,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload image file: %v", err)
		}
	}

	createUserResult, err := as.UserService.CreateUser(ctx, userDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := auth.GenerateAccessToken(createUserResult.UserID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(createUserResult.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	err = as.AuthRepo.InsertRefreshToken(ctx, mapper.ToRefreshTokenEntity(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &authDomain.AuthTokensAndUser{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		User: domain.User{
			FullName:        userDomain.FullName,
			Username:        userDomain.Username,
			Email:           userDomain.Email,
			ProfileImageURL: userDomain.ProfileImageURL,
		},
	}, nil
}

func (as *authService) Login(ctx context.Context, loginInput authDomain.Login) (*authDomain.AuthTokensAndUser, error) {
	userDetails, err := as.UserService.AuthenticateUser(ctx, loginInput.Email, loginInput.Password)
	if err != nil {
		return nil, fmt.Errorf("failed authenticating user: %w", err)
	}

	as.logger.Info("Authentication successful. Generating tokens...", zap.String("userID", userDetails.ID))

	// Revoke/delete other user associated refresh tokens
	err = as.AuthRepo.RevokeAllRefreshTokens(ctx, userDetails.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	accessToken, err := auth.GenerateAccessToken(userDetails.ID, []byte(as.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(userDetails.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenEntity := mapper.ToRefreshTokenEntity(refreshToken)

	err = as.AuthRepo.InsertRefreshToken(ctx, refreshTokenEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &authDomain.AuthTokensAndUser{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		User: domain.User{
			FullName: userDetails.FullName,
			Username: userDetails.Username,
			Email:    userDetails.Email,
		},
	}, nil
}

func (as *authService) RefreshToken(ctx context.Context, refreshInput authDomain.Refresh) (*authDomain.AuthTokens, error) {
	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, refreshInput.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, err
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
		return nil, err
	}

	newRefreshToken, err := auth.GenerateRefreshToken(refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newRefreshTokenEntity := mapper.ToRefreshTokenEntity(newRefreshToken)

	err = as.AuthRepo.InsertRefreshToken(ctx, newRefreshTokenEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return &authDomain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
	}, nil
}

func (as *authService) RevokeRefreshToken(ctx context.Context, revokeRefreshTokenInput authDomain.RevokeRefreshToken) error {
	as.logger.Info("Refreshing tokens...")

	refreshToken, err := as.AuthRepo.GetRefreshToken(ctx, revokeRefreshTokenInput.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRefreshTokenNotFound
		}

		return err
	}

	if refreshToken.Revoked {
		return ErrRefreshTokenAlreadyRevoked
	}

	err = as.AuthRepo.RevokeRefreshToken(ctx, refreshToken.ID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}
