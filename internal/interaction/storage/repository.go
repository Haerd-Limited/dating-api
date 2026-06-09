package storage

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	CountSuperlikesSince(ctx context.Context, userID string, since time.Time, exec boil.ContextExecutor) (int64, error)
	CountActiveMatches(ctx context.Context, userID string, exec boil.ContextExecutor) (int64, error)
	CountActiveMatchesForUsers(ctx context.Context, userIDs []string) (map[string]int64, error)
	LockUsersForMatchCreation(ctx context.Context, tx *sql.Tx, userA, userB string) error
	ListSwipesByUserID(ctx context.Context, userID string) ([]*entity.Swipe, error)
	HasIncomingLike(ctx context.Context, watcherID, likerID string) (bool, error)
	InsertWatch(ctx context.Context, watcherID, watchedID string) error
	DeleteWatch(ctx context.Context, watcherID, watchedID string) error
	DeleteWatchBetween(ctx context.Context, userA, userB string) error
	GetWatchedUserIDs(ctx context.Context, watcherID string) (map[string]struct{}, error)
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

func (is *repository) ListSwipesByUserID(ctx context.Context, userID string) ([]*entity.Swipe, error) {
	return entity.Swipes(
		qm.Where("(actor_id = ? OR target_id = ?)", userID, userID),
		qm.OrderBy(entity.SwipeColumns.CreatedAt+" ASC"),
	).All(ctx, is.db)
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
		qm.Where("(user_a = ? OR user_b = ?) AND status = ?", userID, userID, string(entity.MatchStatusActive)),
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

func (is *repository) CountSuperlikesSince(ctx context.Context, userID string, since time.Time, exec boil.ContextExecutor) (int64, error) {
	if exec == nil {
		exec = is.db
	}

	count, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(userID),
		entity.SwipeWhere.Action.EQ(constants.ActionSuperlike),
		entity.SwipeWhere.CreatedAt.GTE(since),
	).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count superlikes since userID=%s since=%s: %w", userID, since.Format(time.RFC3339), err)
	}

	return count, nil
}

func (is *repository) CountActiveMatches(ctx context.Context, userID string, exec boil.ContextExecutor) (int64, error) {
	if exec == nil {
		exec = is.db
	}

	count, err := entity.Matches(
		qm.Where("(user_a = ? OR user_b = ?) AND status = ?", userID, userID, string(entity.MatchStatusActive)),
	).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count active matches userID=%s: %w", userID, err)
	}

	return count, nil
}

// CountActiveMatchesForUsers returns a map of userID -> active match count for
// the given user IDs. Missing keys mean zero. The caller should treat absent
// entries as 0.
func (is *repository) CountActiveMatchesForUsers(ctx context.Context, userIDs []string) (map[string]int64, error) {
	if len(userIDs) == 0 {
		return map[string]int64{}, nil
	}

	const query = `
		SELECT user_id, COUNT(*)::bigint AS active_count
		FROM (
			SELECT user_a AS user_id FROM matches
			WHERE status = 'active' AND user_a = ANY($1)
			UNION ALL
			SELECT user_b AS user_id FROM matches
			WHERE status = 'active' AND user_b = ANY($1)
		) t
		GROUP BY user_id
	`

	rows, err := is.db.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("query active matches for users: %w", err)
	}

	defer func() { _ = rows.Close() }()

	counts := make(map[string]int64, len(userIDs))

	for rows.Next() {
		var (
			id    string
			count int64
		)

		if err := rows.Scan(&id, &count); err != nil {
			return nil, fmt.Errorf("scan active matches row: %w", err)
		}

		counts[id] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active matches rows: %w", err)
	}

	return counts, nil
}

// LockUsersForMatchCreation acquires two transaction-scoped advisory locks
// (one per user) inside the given tx, sorted to avoid deadlocks. Use
// hashtextextended (PG 13+) which returns int8 matching pg_advisory_xact_lock's
// signature without a cast.
func (is *repository) LockUsersForMatchCreation(ctx context.Context, tx *sql.Tx, userA, userB string) error {
	a, b := userA, userB
	if b < a {
		a, b = b, a
	}

	_, err := tx.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0)),
		        pg_advisory_xact_lock(hashtextextended($2, 0))`,
		a, b,
	)
	if err != nil {
		return fmt.Errorf("acquire advisory locks for match creation userA=%s userB=%s: %w", a, b, err)
	}

	return nil
}

func (is *repository) HasIncomingLike(ctx context.Context, watcherID, likerID string) (bool, error) {
	ok, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(likerID),
		entity.SwipeWhere.TargetID.EQ(watcherID),
		qm.Where("action IN (?, ?)", constants.ActionLike, constants.ActionSuperlike),
	).Exists(ctx, is.db)
	if err != nil {
		return false, fmt.Errorf("check incoming like watcherID=%s likerID=%s: %w", watcherID, likerID, err)
	}

	return ok, nil
}

func (is *repository) InsertWatch(ctx context.Context, watcherID, watchedID string) error {
	const stmt = `
INSERT INTO match_slot_watches (watcher_user_id, watched_user_id)
VALUES ($1, $2)
ON CONFLICT (watcher_user_id, watched_user_id) DO NOTHING`

	_, err := is.db.ExecContext(ctx, stmt, watcherID, watchedID)
	if err != nil {
		return fmt.Errorf("insert watch watcherID=%s watchedID=%s: %w", watcherID, watchedID, err)
	}

	return nil
}

func (is *repository) DeleteWatch(ctx context.Context, watcherID, watchedID string) error {
	const stmt = `DELETE FROM match_slot_watches WHERE watcher_user_id = $1 AND watched_user_id = $2`

	_, err := is.db.ExecContext(ctx, stmt, watcherID, watchedID)
	if err != nil {
		return fmt.Errorf("delete watch watcherID=%s watchedID=%s: %w", watcherID, watchedID, err)
	}

	return nil
}

func (is *repository) DeleteWatchBetween(ctx context.Context, userA, userB string) error {
	const stmt = `
DELETE FROM match_slot_watches
WHERE (watcher_user_id = $1 AND watched_user_id = $2)
   OR (watcher_user_id = $2 AND watched_user_id = $1)`

	_, err := is.db.ExecContext(ctx, stmt, userA, userB)
	if err != nil {
		return fmt.Errorf("delete watch between userA=%s userB=%s: %w", userA, userB, err)
	}

	return nil
}

func (is *repository) GetWatchedUserIDs(ctx context.Context, watcherID string) (map[string]struct{}, error) {
	const stmt = `SELECT watched_user_id FROM match_slot_watches WHERE watcher_user_id = $1`

	rows, err := is.db.QueryContext(ctx, stmt, watcherID)
	if err != nil {
		return nil, fmt.Errorf("get watched user ids watcherID=%s: %w", watcherID, err)
	}

	defer func() { _ = rows.Close() }()

	watched := make(map[string]struct{})

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan watched user id: %w", err)
		}

		watched[id] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate watched user ids: %w", err)
	}

	return watched, nil
}
