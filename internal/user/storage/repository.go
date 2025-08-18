package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type UserRepository interface {
	InsertUser(ctx context.Context, user *entity.User) (*string, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	FollowUser(ctx context.Context, input *domain.FollowUser) error
	UnfollowUser(ctx context.Context, input *domain.UnfollowUser) error
	GetUsersFollowing(ctx context.Context, userID string) ([]*entity.User, error)
	GetUsersFollowers(ctx context.Context, userID string) ([]*entity.User, error)
	CountFollowers(ctx context.Context, userID string) (int, error)
	CountFollowing(ctx context.Context, userID string) (int, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
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
	user, err := entity.Users(entity.UserWhere.Email.EQ(email)).One(ctx, r.db)
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

func (r *userRepository) FollowUser(ctx context.Context, input *domain.FollowUser) error {
	follow := &entity.Follow{
		FollowerID:  input.FollowerID,
		FollowingID: input.FollowingID,
	}

	err := follow.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		// If already exists, return nil (idempotent behavior)
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil // unique_violation
		}

		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

func (r *userRepository) UnfollowUser(ctx context.Context, input *domain.UnfollowUser) error {
	// Build delete query
	_, err := entity.Follows(
		qm.Where("follower_id = ? AND following_id = ?", input.FollowerID, input.FollowingID),
	).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}

	return nil
}

func (r *userRepository) GetUsersFollowing(ctx context.Context, userID string) ([]*entity.User, error) {
	// Get all follow rows where this user is the follower
	follows, err := entity.Follows(
		qm.Where("follower_id = ?", userID),
		qm.Load(entity.FollowRels.Following), // Load the user they're following
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get users following: %w", err)
	}

	var users []*entity.User

	for _, f := range follows {
		if f.R != nil && f.R.Following != nil {
			users = append(users, &entity.User{
				ID:          f.R.Following.ID,
				Username:    f.R.Following.Username,
				FullName:    f.R.Following.FullName,
				Email:       f.R.Following.Email,
				Bio:         f.R.Following.Bio,
				Gender:      f.R.Following.Gender,
				DateOfBirth: f.R.Following.DateOfBirth,
				CreatedAt:   f.R.Following.CreatedAt,
			})
		}
	}

	return users, nil
}

func (r *userRepository) GetUsersFollowers(ctx context.Context, userID string) ([]*entity.User, error) {
	// Get all follow rows where this user is being followed
	follows, err := entity.Follows(
		qm.Where("following_id = ?", userID),
		qm.Load(entity.FollowRels.Follower), // Load the follower user
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	var users []*entity.User

	for _, f := range follows {
		if f.R != nil && f.R.Follower != nil {
			users = append(users, &entity.User{
				ID:          f.R.Follower.ID,
				Username:    f.R.Follower.Username,
				FullName:    f.R.Follower.FullName,
				Email:       f.R.Follower.Email,
				Bio:         f.R.Follower.Bio,
				Gender:      f.R.Follower.Gender,
				DateOfBirth: f.R.Follower.DateOfBirth,
				CreatedAt:   f.R.Follower.CreatedAt,
			})
		}
	}

	return users, nil
}

func (r *userRepository) CountFollowers(ctx context.Context, userID string) (int, error) {
	count, err := entity.Follows(
		entity.FollowWhere.FollowingID.EQ(userID),
	).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count followers: %w", err)
	}

	return int(count), nil
}

func (r *userRepository) CountFollowing(ctx context.Context, userID string) (int, error) {
	count, err := entity.Follows(
		entity.FollowWhere.FollowerID.EQ(userID),
	).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count following: %w", err)
	}

	return int(count), nil
}

func (r *userRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	exists, err := entity.Follows(
		entity.FollowWhere.FollowerID.EQ(followerID),
		entity.FollowWhere.FollowingID.EQ(followeeID),
	).Exists(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("failed to check following status: %w", err)
	}

	return exists, nil
}
