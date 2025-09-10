package storage

import (
	"context"
	"github.com/friendsofgo/errors"
	"github.com/lib/pq"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type InteractionRepository interface {
	InsertSwipe(ctx context.Context, swipe entity.Swipe) error
}

type repository struct {
	db *sqlx.DB
}

func NewInteractionRepository(db *sqlx.DB) InteractionRepository {
	return &repository{
		db: db,
	}
}

var (
	ErrAlreadySwiped = errors.New("you've already swiped on this user.")
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
