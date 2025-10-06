package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/friendsofgo/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type InteractionRepository interface {
	InsertSwipe(ctx context.Context, swipe entity.Swipe, tx *sql.Tx) error
	CheckIfMatchable(ctx context.Context, userID string, targetUserID string) (bool, error)
	CreateMatch(ctx context.Context, match entity.Match, tx *sql.Tx) error
	GetIncomingLikes(ctx context.Context, userID string, limit, offset int) ([]string, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
	AlreadyMatched(ctx context.Context, userID string, targetUserID string) (bool, error)
	GetSwipeByActorIDAndTargetID(ctx context.Context, actorID, targetID string) (*entity.Swipe, error)
	GetFirstLikeSwipeByBetweenUsers(ctx context.Context, userA, userB string) (*entity.Swipe, error)
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

func (is *repository) GetFirstLikeSwipeByBetweenUsers(ctx context.Context, userA, userB string) (*entity.Swipe, error) {
	s, err := entity.Swipes(
		// (A -> B) OR (B -> A)
		qm.Where("(actor_id = ? AND target_id = ?) OR (actor_id = ? AND target_id = ?)",
			userA, userB, userB, userA),
		// action in ('like', 'superlike')
		qm.Where("action IN (?, ?)", constants.ActionLike, constants.ActionSuperlike),
		qm.OrderBy("created_at ASC"),
		qm.Limit(1),
	).One(ctx, is.db)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (is *repository) GetSwipeByActorIDAndTargetID(ctx context.Context, actorID, targetID string) (*entity.Swipe, error) {
	s, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(actorID),
		entity.SwipeWhere.TargetID.EQ(targetID),
	).One(ctx, is.db)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// GetIncomingLikes returns profiles of users who have liked/superliked `userID`.
func (is *repository) GetIncomingLikes(ctx context.Context, userID string, limit, offset int) ([]string, error) {
	swipes, err := entity.Swipes(
		entity.SwipeWhere.TargetID.EQ(userID),
		qm.Where("action IN (?, ?)", constants.ActionLike, constants.ActionSuperlike),
		qm.Where(`NOT EXISTS (
		SELECT 1 FROM swipes s2 
		WHERE s2.actor_id = ? 
		AND s2.target_id = swipes.actor_id 
		AND s2.action = ?
	)`, userID, constants.ActionPass),

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

func (is *repository) InsertSwipe(ctx context.Context, swipe entity.Swipe, tx *sql.Tx) error {
	err := swipe.Insert(ctx, tx, boil.Infer())
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
		qm.Where("action IN (?, ?)", constants.ActionLike, constants.ActionSuperlike),
	).Exists(ctx, is.db)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	return true, nil
}

func (is *repository) CreateMatch(ctx context.Context, match entity.Match, tx *sql.Tx) error {
	err := match.Insert(ctx, tx, boil.Infer())
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
		qm.Where("(user_a = ? AND user_b = ?) OR (user_a = ? AND user_b = ?) ", userID, targetUserID, targetUserID, userID),
	).Exists(ctx, is.db)
	if err != nil {
		return false, err
	}

	if ok {
		return true, nil
	}

	return false, nil
}
