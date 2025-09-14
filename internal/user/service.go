package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/internal/user/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=user
type Service interface {
	CreateUser(ctx context.Context, user domain.User) (userID *string, err error)
	AuthenticateUser(ctx context.Context, phoneNumber string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UserExistsByIdentifier(ctx context.Context, channel, identifier string) (bool, error)
}

type userService struct {
	logger     *zap.Logger
	userRepo   storage.UserRepository
	awsService aws.Service
	cache      *freecache.Cache
}

func NewUserService(
	logger *zap.Logger,
	userRepo storage.UserRepository,
	awsService aws.Service,
	cache *freecache.Cache,
) Service {
	return &userService{
		logger:     logger,
		userRepo:   userRepo,
		awsService: awsService,
		cache:      cache,
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (us *userService) CreateUser(ctx context.Context, user domain.User) (userID *string, err error) {
	userID, err = us.userRepo.InsertUser(ctx, mapper.ToUserEntity(user))
	if err != nil {
		return nil, err
	}

	return userID, nil
}

// AuthenticateUser checks credentials and returns the user if valid, otherwise an error.
func (us *userService) AuthenticateUser(ctx context.Context, phoneNumber string) (*domain.User, error) {
	userEntity, err := us.userRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// todo: update/implement using phonenumber like hinge

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
	updatedEntity, cols := mapper.ToUpdatedUserEntity(*user)
	return us.userRepo.UpdateUser(ctx, updatedEntity, cols)
}

func (us *userService) UserExistsByIdentifier(ctx context.Context, channel, identifier string) (bool, error) {
	switch channel {
	case "sms":
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
