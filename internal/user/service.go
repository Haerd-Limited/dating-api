package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/coocood/freecache"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/internal/user/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=user
type Service interface {
	CreateUser(ctx context.Context, user domain.User) (*string, error)
	AuthenticateUser(ctx context.Context, phoneNumber string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
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

var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUserDetailsAlreadyExists = errors.New("user details already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
)

func (us *userService) CreateUser(ctx context.Context, user domain.User) (*string, error) {
	userID, err := us.userRepo.InsertUser(ctx, mapper.ToUserEntity(user))
	if err != nil { // todo: move to repo layer
		// Check if the error is a unique constraint violation (email already exists)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_email_key":
				return nil, ErrEmailAlreadyExists
			default:
				return nil, ErrUserDetailsAlreadyExists
			}
		}

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
	return us.userRepo.UpdateUser(ctx, mapper.ToUserEntity(*user))
}
