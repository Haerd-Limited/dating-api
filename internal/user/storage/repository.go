package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type UserRepository interface {
	InsertUser(ctx context.Context, user *entity.User, tx *sql.Tx) (string, error)
	GetByPhoneNumber(ctx context.Context, number string) (*entity.User, error)
	CheckUserExistenceByPhoneNumber(ctx context.Context, number string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	UpdateUser(ctx context.Context, e *entity.User, cols []string) error
	DeleteUser(ctx context.Context, userID string) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

var (
	ErrUserDoesNotExists        = errors.New("user does not exists")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUserDetailsAlreadyExists = errors.New("user details already exists")
)

func (r *userRepository) UpdateUser(ctx context.Context, e *entity.User, cols []string) error {
	if len(cols) == 0 {
		return nil
	} // nothing to change

	_, err := e.Update(ctx, r.db, boil.Whitelist(cols...))
	if err != nil {
		// Check if the error is a unique constraint violation (email already exists)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_email_key":
				return ErrEmailAlreadyExists
			default:
				return ErrUserDetailsAlreadyExists
			}
		}

		return fmt.Errorf("failed to update user %s: %w", e.ID, err)
	}

	return nil
}

func (r *userRepository) InsertUser(ctx context.Context, user *entity.User, tx *sql.Tx) (string, error) {
	err := user.Insert(ctx, tx, boil.Infer())
	if err != nil {
		// Check if the error is a unique constraint violation (email already exists)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_email_key":
				return "", ErrEmailAlreadyExists
			default:
				return "", ErrUserDetailsAlreadyExists
			}
		}

		return "", fmt.Errorf("failed inserting user entity: %w", err)
	}

	// If ID is DB-generated (DEFAULT gen_random_uuid()), make sure we have it loaded
	if user.ID == "" {
		if err = user.Reload(ctx, tx); err != nil {
			return "", fmt.Errorf("reload user after insert: %w", err)
		}
	}

	return user.ID, nil
}

// GetByPhoneNumber retrieves a user by their phoneNumber.
func (r *userRepository) GetByPhoneNumber(ctx context.Context, number string) (*entity.User, error) {
	user, err := entity.Users(entity.UserWhere.Phone.EQ(null.StringFrom(number))).One(ctx, r.db)
	if err != nil {
		return nil, err
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

func (r *userRepository) CountUsers(ctx context.Context) (int64, error) {
	n, err := entity.Users().Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return n, nil
}

func (r *userRepository) CheckUserExistenceByPhoneNumber(ctx context.Context, number string) (bool, error) {
	exists, err := entity.Users(entity.UserWhere.Phone.EQ(null.StringFrom(number))).Exists(ctx, r.db)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// DeleteUser deletes a user by ID. This will cascade delete all related records.
func (r *userRepository) DeleteUser(ctx context.Context, userID string) error {
	_, err := entity.Users(entity.UserWhere.ID.EQ(userID)).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", userID, err)
	}

	return nil
}
