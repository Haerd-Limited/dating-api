package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/coocood/freecache"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	awsDomain "github.com/Haerd-Limited/dating-api/internal/aws/domain"

	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/internal/user/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=user
type Service interface {
	CreateUser(ctx context.Context, user *domain.User) (*CreateUserResult, error)
	AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error)
	GetUserDetails(ctx context.Context, id string) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	UpdateUserProfile(ctx context.Context, input *domain.UpdateUserProfile) (*domain.User, error)
}

type userService struct {
	logger     *zap.Logger
	userRepo   storage.UserRepository
	awsService aws.AWSService
	cache      *freecache.Cache
}

func NewUserService(
	logger *zap.Logger,
	userRepo storage.UserRepository,
	awsService aws.AWSService,
	cache *freecache.Cache,
) Service {
	return &userService{
		logger:     logger,
		userRepo:   userRepo,
		awsService: awsService,
		cache:      cache,
	}
}

type CreateUserResult struct {
	UserID string
}

var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUserNameAlreadyExists    = errors.New("username already exists")
	ErrUserDetailsAlreadyExists = errors.New("user details already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
)

func (us *userService) CreateUser(ctx context.Context, user *domain.User) (*CreateUserResult, error) {
	userID, err := us.userRepo.InsertUser(ctx, mapper.ToUserEntity(user))
	if err != nil {
		// Check if the error is a unique constraint violation (email already exists)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, ErrUserNameAlreadyExists
			case "users_email_key":
				return nil, ErrEmailAlreadyExists
			default:
				return nil, ErrUserDetailsAlreadyExists
			}
		}

		return nil, err
	}

	return &CreateUserResult{
		UserID: *userID,
	}, nil
}

// AuthenticateUser checks credentials and returns the user if valid, otherwise an error.
func (us *userService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := us.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	userDomain := mapper.UserEntityToUserDomain(user)

	return userDomain, nil
}

func (us *userService) GetUserDetails(ctx context.Context, id string) (*domain.User, error) {
	userEntity, err := us.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id(%s): %w", id, err)
	}

	user := mapper.UserEntityToUserDomain(userEntity)

	return user, nil
}

func (us *userService) GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	var users []*domain.User

	for _, id := range ids {
		userEntity, err := us.userRepo.GetUserByID(ctx, id)
		if err != nil {
			// Optionally log or skip on error per user instead of failing the whole batch
			us.logger.Sugar().Errorw("failed to fetch user", "userID", id, "error", err)
			continue
		}

		user := mapper.UserEntityToUserDomain(userEntity)
		users = append(users, user)
	}

	return users, nil
}

func (us *userService) UpdateUserProfile(ctx context.Context, input *domain.UpdateUserProfile) (*domain.User, error) {
	us.logger.Info("Updating user profile...", zap.String("userID", input.UserID))

	// Step 1: Fetch user from database
	existingUser, err := us.userRepo.GetUserByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user: %w", err)
	}
	// TODO: Implement Andre advice - convert to domain first and then update with domain, then convert to entity again
	// Should be domain business logic on DOMAIN objects. not entity

	// Step 2: Update fields if provided
	if input.FullName != "" {
		existingUser.FullName = input.FullName
	}

	if input.Username != "" {
		existingUser.Username = input.Username
	}

	if input.Email != "" {
		existingUser.Email = input.Email
	}

	if input.DateOfBirth != nil {
		existingUser.DateOfBirth = null.TimeFromPtr(input.DateOfBirth)
	}

	if input.Gender != "" {
		existingUser.Gender = null.StringFrom(input.Gender)
	}

	if input.Biography != "" {
		existingUser.Bio = null.StringFrom(input.Biography)
	}

	// Step 3: Handle profile image upload if provided
	if input.ProfileImage != nil && input.ImageHeader != nil {
		// upload post image to s3
		var oldProfileImageURL string
		if existingUser.ProfilePicURL.Valid {
			oldProfileImageURL = existingUser.ProfilePicURL.String
		}

		newProfileImageURL, err := us.awsService.UploadImage(ctx, awsDomain.ImageUpload{
			AuthorID:    existingUser.Username,
			ImageHeader: *input.ImageHeader,
			ImageFile:   *input.ProfileImage,
			FolderPath:  awsDomain.FolderProfilePictures,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload image file: %v", err)
		}

		existingUser.ProfilePicURL = null.StringFromPtr(newProfileImageURL)

		// delete old profile image if it exists from bucket
		if oldProfileImageURL != "" {
			err = us.awsService.DeleteFile(ctx, oldProfileImageURL)
			if err != nil {
				us.logger.Warn("failed to delete old image file", zap.Error(err))
			}
		}
	}

	// Step 4: Persist updated user
	if err = us.userRepo.UpdateUser(ctx, existingUser); err != nil {
		// Check if the error is a unique constraint violation (email already exists)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, ErrUserNameAlreadyExists
			case "users_email_key":
				return nil, ErrEmailAlreadyExists
			default:
				return nil, ErrUserDetailsAlreadyExists
			}
		}

		return nil, fmt.Errorf("failed updating user: %w", err)
	}

	// Step 5: Map to domain and return
	return mapper.UserEntityToUserDomain(existingUser), nil
}
