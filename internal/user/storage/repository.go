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
	GetByPhoneNumber(ctx context.Context, number string) (*entity.User, error)
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
	// Start a transaction so user and scaffold are atomic
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	rollback := func(e error) (*string, error) {
		_ = tx.Rollback()
		return nil, e
	}

	// 1) Insert user
	err = user.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return rollback(fmt.Errorf("failed inserting user entity: %w", err))
	}

	// If ID is DB-generated (DEFAULT gen_random_uuid()), make sure we have it loaded
	if user.ID == "" {
		if err = user.Reload(ctx, tx); err != nil {
			return rollback(fmt.Errorf("reload user after insert: %w", err))
		}
	}

	// 2) Scaffold empty profile
	profile := &entity.UserProfile{
		UserID:      user.ID,
		DisplayName: null.StringFrom(fmt.Sprintf("%s %s", user.FirstName, user.LastName.String)),
	}

	err = profile.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return rollback(fmt.Errorf("insert user_profile: %w", err))
	}

	// 3) Scaffold preferences row (timestamps have DEFAULT now())
	prefs := &entity.UserPreference{
		UserID: user.ID,
	}

	err = prefs.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return rollback(fmt.Errorf("insert user_preferences: %w", err))
	}

	// 4) Commit
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &user.ID, nil
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
