package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type Repository interface {
	GetOutgoingLikeTargetIDs(ctx context.Context, actorID string) ([]string, error)
	GetFavouritersOfUser(ctx context.Context, watchedUserID string) ([]string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetOutgoingLikeTargetIDs(ctx context.Context, actorID string) ([]string, error) {
	const stmt = `
SELECT s.target_id
  FROM swipes s
 WHERE s.actor_id = $1
   AND s.action IN ('` + constants.ActionLike + `', '` + constants.ActionSuperlike + `')
   AND NOT EXISTS (
       SELECT 1 FROM matches m
        WHERE (m.user_a = s.actor_id AND m.user_b = s.target_id)
           OR (m.user_a = s.target_id AND m.user_b = s.actor_id)
   )
   AND NOT EXISTS (
       SELECT 1 FROM user_blocks b
        WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = s.target_id)
           OR (b.blocker_user_id = s.target_id AND b.blocked_user_id = $1)
   )`

	rows, err := r.db.QueryContext(ctx, stmt, actorID)
	if err != nil {
		return nil, fmt.Errorf("get outgoing like targets actorID=%s: %w", actorID, err)
	}

	defer func() { _ = rows.Close() }()

	var targets []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan outgoing like target: %w", err)
		}

		targets = append(targets, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outgoing like targets: %w", err)
	}

	return targets, nil
}

func (r *repository) GetFavouritersOfUser(ctx context.Context, watchedUserID string) ([]string, error) {
	const stmt = `SELECT watcher_user_id FROM match_slot_watches WHERE watched_user_id = $1`

	rows, err := r.db.QueryContext(ctx, stmt, watchedUserID)
	if err != nil {
		return nil, fmt.Errorf("get favouriters watchedUserID=%s: %w", watchedUserID, err)
	}

	defer func() { _ = rows.Close() }()

	var favouriters []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan favouriter: %w", err)
		}

		favouriters = append(favouriters, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favouriters: %w", err)
	}

	return favouriters, nil
}
