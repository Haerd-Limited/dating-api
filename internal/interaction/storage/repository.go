package storage

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/friendsofgo/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type InteractionRepository interface {
	InsertSwipe(ctx context.Context, swipe entity.Swipe) error
	CheckIfMatchable(ctx context.Context, userID string, targetUserID string) (bool, error)
	CreateMatch(ctx context.Context, match entity.Match) error
	GetIncomingLikes(ctx context.Context, userID string, limit, offset int) ([]string, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
	AlreadyMatched(ctx context.Context, userID string, targetUserID string) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewInteractionRepository(db *sqlx.DB) InteractionRepository {
	return &repository{
		db: db,
	}
}

var ErrAlreadySwiped = errors.New("you've already swiped on this user.")

const (
	actionLike      = "like"
	actionPass      = "pass"
	actionSuperlike = "superlike"
)

// GetIncomingLikes returns profiles of users who have liked/superliked `userID`.
func (is *repository) GetIncomingLikes(ctx context.Context, userID string, limit, offset int) ([]string, error) {
	swipes, err := entity.Swipes(
		entity.SwipeWhere.TargetID.EQ(userID),
		qm.Where("action IN (?, ?)", actionLike, actionSuperlike),
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy("created_at ASC"),
	).All(ctx, is.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incoming likes for userID=%s: %w", userID, err)
	}

	var userIDs []string
	for _, s := range swipes {
		userIDs = append(userIDs, s.ActorID) // the "liker"
	}

	return userIDs, nil
}

func (is *repository) InsertSwipe(ctx context.Context, swipe entity.Swipe) error {
	err := swipe.Insert(ctx, is.db, boil.Infer())
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "swipes_actor_target_uniq":
				return ErrAlreadySwiped
			}

			return err
		}

		return err
	}

	return nil
}

func (is *repository) CheckIfMatchable(ctx context.Context, userID string, targetUserID string) (bool, error) {
	ok, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(targetUserID), // they swiped
		entity.SwipeWhere.TargetID.EQ(userID),      // on me
		qm.Where("action IN (?, ?)", actionLike, actionSuperlike),
	).Exists(ctx, is.db)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	return true, nil
}

func (is *repository) CreateMatch(ctx context.Context, match entity.Match) error {
	err := match.Insert(ctx, is.db, boil.Infer())
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// already matched – treat as success
			return nil
		}

		return err
	}

	return nil
}

func (is *repository) GetMatches(ctx context.Context, userID string) ([]*entity.Match, error) {
	matches, err := entity.Matches(
		entity.MatchWhere.UserB.EQ(userID),
		qm.Or2(
			entity.MatchWhere.UserA.EQ(userID),
		),
	).All(ctx, is.db)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (is *repository) AlreadyMatched(ctx context.Context, userID string, targetUserID string) (bool, error) {
	ok, err := entity.Matches(
		entity.MatchWhere.UserA.EQ(userID),
		entity.MatchWhere.UserB.EQ(targetUserID),
	).Exists(ctx, is.db)
	if err != nil {
		return false, err
	}

	if ok {
		return true, nil
	}

	return false, nil
}
