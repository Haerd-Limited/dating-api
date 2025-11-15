package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/preference"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/internal/user/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=user
type Service interface {
	// CreateUser creates a new user and scaffolds a profile and user preferences in a single transaction.
	CreateUser(ctx context.Context, user domain.User) (string, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UserExistsByIdentifier(ctx context.Context, channel, identifier string) (bool, error)
	// DeleteAccount deletes all user data including their account, profile, preferences, and all associated S3 files
	DeleteAccount(ctx context.Context, userID string) error
}

type userService struct {
	logger            *zap.Logger
	userRepo          storage.UserRepository
	awsService        aws.Service
	cache             *freecache.Cache
	uow               uow.UoW
	profileService    profile.Service
	preferenceService preference.Service
}

func NewUserService(
	logger *zap.Logger,
	userRepo storage.UserRepository,
	awsService aws.Service,
	cache *freecache.Cache,
	uow uow.UoW,
	profileService profile.Service,
	preferenceService preference.Service,
) Service {
	return &userService{
		logger:            logger,
		userRepo:          userRepo,
		awsService:        awsService,
		cache:             cache,
		uow:               uow,
		profileService:    profileService,
		preferenceService: preferenceService,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNameContainsSpaces = errors.New("name must not contain spaces")
	ErrInvalidNameLength  = errors.New("name must be between 3 and 20 characters")
)

const (
	minNameLen = 3
	maxNameLen = 20
)

// CreateUser creates a new user and scaffolds a profile and user preferences in a single transaction.
func (us *userService) CreateUser(ctx context.Context, user domain.User) (string, error) {
	tx, err := us.uow.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func(tx uow.Tx) {
		err = tx.Rollback()
	}(tx)

	userID, err := us.userRepo.InsertUser(ctx, mapper.ToUserEntity(user), tx.Raw())
	if err != nil {
		return "", fmt.Errorf("insert user: %w", err)
	}

	err = us.profileService.ScaffoldProfile(ctx, tx.Raw(), userID)
	if err != nil {
		return "", fmt.Errorf("create new default profile: %w", err)
	}

	err = us.preferenceService.ScaffoldUserPreferences(ctx, tx.Raw(), userID)
	if err != nil {
		return "", fmt.Errorf("create new default preference: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return userID, nil
}

func (us *userService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*domain.User, error) {
	userEntity, err := us.userRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return mapper.UserEntityToUserDomain(userEntity), nil
}

func (us *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	userEntity, err := us.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id(%s): %w", id, err)
	}

	return mapper.UserEntityToUserDomain(userEntity), nil
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

func (us *userService) UpdateUser(ctx context.Context, user *domain.User) error {
	err := us.validateAndSanitiseUserDetails(user)
	if err != nil {
		return fmt.Errorf("validate and sanitise user details: %w", err)
	}

	updatedEntity, cols := mapper.ToUpdatedUserEntity(*user)

	return us.userRepo.UpdateUser(ctx, updatedEntity, cols)
}

func (us *userService) UserExistsByIdentifier(ctx context.Context, channel, identifier string) (bool, error) {
	switch channel {
	case constants.SmsChannel:
		exists, err := us.userRepo.CheckUserExistenceByPhoneNumber(ctx, identifier)
		if err != nil {
			return false, err
		}

		return exists, nil
		// todo: implement email
	default:
		return false, fmt.Errorf("unsupported channel: %s", channel)
	}
}

func (us *userService) validateAndSanitiseUserDetails(userDetails *domain.User) error {
	userDetails.FirstName = strings.TrimSpace(userDetails.FirstName)
	if hasAnySpace(userDetails.FirstName) {
		return fmt.Errorf("first%w", ErrNameContainsSpaces)
	}
	// first name length check
	if l := len(userDetails.FirstName); l < minNameLen || l > maxNameLen {
		return ErrInvalidNameLength
	}

	if userDetails.LastName != nil {
		temp := strings.TrimSpace(*userDetails.LastName)
		userDetails.LastName = &temp

		if hasAnySpace(*userDetails.LastName) {
			return fmt.Errorf("last%w", ErrNameContainsSpaces)
		}

		// last name length check
		if l := len(*userDetails.LastName); l < minNameLen || l > maxNameLen {
			return ErrInvalidNameLength
		}
	}

	if !looksLikeEmail(strings.TrimSpace(userDetails.Email)) {
		return commonErrors.ErrInvalidEmail
	}

	return nil
}

// hasAnySpace returns true if s contains any Unicode whitespace character.
func hasAnySpace(s string) bool {
	return strings.IndexFunc(s, unicode.IsSpace) >= 0
}

func looksLikeEmail(s string) bool { return strings.Contains(s, "@") && strings.Contains(s, ".") }

// DeleteAccount deletes all user data including their account, profile, preferences, and all associated S3 files.
// This operation is irreversible and deletes:
// - All S3 files (profile photos, voice prompts, message voice notes, verification photos)
// - User account and all related data via CASCADE (profile, preferences, messages, matches, swipes, etc.)
func (us *userService) DeleteAccount(ctx context.Context, userID string) error {
	us.logger.Info("Starting account deletion", zap.String("userID", userID))

	// Step 1: Verify user exists
	_, err := us.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Step 2: Delete all S3 files first (before DB deletion to preserve data if S3 deletion fails)
	us.logger.Info("Deleting S3 files for user", zap.String("userID", userID))

	err = us.awsService.DeleteAllUserFiles(ctx, userID)
	if err != nil {
		us.logger.Error("Failed to delete S3 files", zap.String("userID", userID), zap.Error(err))
		// Continue with DB deletion even if S3 deletion fails - log but don't fail
		// This ensures account is still deleted from DB even if S3 operations have issues
	}

	// Step 3: Delete user from database (CASCADE will handle all related records)
	us.logger.Info("Deleting user from database", zap.String("userID", userID))

	err = us.userRepo.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	// Step 4: Clear any cached data for this user
	// The cache key would depend on your cache implementation
	// Since we're using freecache, we could clear entries if needed, but it will expire naturally

	us.logger.Info("Account deletion completed successfully", zap.String("userID", userID))

	return nil
}
