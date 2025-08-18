package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aarondl/null/v8"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type UserRepository interface {
	InsertUser(ctx context.Context, user *entity.User) (*string, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

var ErrUserDoesNotExists = errors.New("user does not exists")

func (r *userRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	_, err := user.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to update user with ID %s: %w", user.ID, err)
	}

	return nil
}

func (r *userRepository) InsertUser(ctx context.Context, user *entity.User) (*string, error) {
	if err := user.Insert(ctx, r.db, boil.Infer()); err != nil {
		return nil, fmt.Errorf("failed inserting user entity: %w", err)
	}

	return &user.ID, nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := entity.Users(entity.UserWhere.Email.EQ(null.StringFrom(email))).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their id
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := entity.Users(entity.UserWhere.ID.EQ(id)).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserDoesNotExists
		}

		return nil, fmt.Errorf("failed to get user by id %s: %w", id, err)
	}

	return user, nil
}
