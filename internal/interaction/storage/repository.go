package storage

import (
	"context"

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
	like      = "like"
	pass      = "pass"
	superlike = "superlike"
)

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
		qm.Where("action IN (?, ?)", like, superlike),
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
