package storage

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/dailypicks/domain"
)

type Repository interface {
	ScoreCandidates(ctx context.Context, userID string, seekGenderIDs []int16, shortlistSize int) ([]domain.ScoredCandidate, error)
	ReplacePendingBatch(ctx context.Context, tx *sql.Tx, userID string, batchDate time.Time, picks []domain.Pick) (string, error)
	GetBatchByState(ctx context.Context, exec boil.ContextExecutor, userID, state string) (*domain.Batch, error)
	GetBatchByID(ctx context.Context, exec boil.ContextExecutor, batchID string) (*domain.Batch, error)
	CountUndecidedPicks(ctx context.Context, exec boil.ContextExecutor, batchID string) (int, error)
	MarkBatchState(ctx context.Context, tx *sql.Tx, batchID, state string, revealedAt *time.Time) error
	MarkPickDecided(ctx context.Context, exec boil.ContextExecutor, userID, targetUserID, status string) error
	ExpirePicks(ctx context.Context, tx *sql.Tx, pickIDs []string) error
	ListEligibleUserIDs(ctx context.Context) ([]string, error)
	LockUserForPicks(ctx context.Context, tx *sql.Tx, userID string) error
	SeekGenderIDs(ctx context.Context, userID string) ([]int16, error)
	CountActiveMatches(ctx context.Context, exec boil.ContextExecutor, userID string) (int64, error)
	HasAnyBatch(ctx context.Context, exec boil.ContextExecutor, userID string) (bool, error)
	IsTargetStillValid(ctx context.Context, viewerID, targetID string) (bool, error)
	InsertPicks(ctx context.Context, tx *sql.Tx, batchID, userID string, picks []domain.Pick) error
	MarkBatchNotified(ctx context.Context, batchIDs []string) error
	ListPendingPickTargetIDs(ctx context.Context, batchID string) ([]string, error)
}

type repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewRepository(db *sqlx.DB, logger *zap.Logger) Repository {
	return &repository{db: db, logger: logger}
}

const scoreCandidatesQuery = `
WITH prefs AS (
    SELECT distance_km, age_min, age_max,
           seek_religion_ids, seek_sexuality_ids, seek_ethnicity_ids
    FROM user_preferences WHERE user_id = $1
),
viewer AS (
    SELECT geo FROM user_profiles WHERE user_id = $1
),
candidates AS (
    SELECT p.user_id, p.birthdate, p.geo, p.religion_id, p.sexuality_id
    FROM user_profiles p
    JOIN users u ON u.id = p.user_id
    WHERE p.user_id <> $1
      AND u.onboarding_step = 'COMPLETE'
      AND u.account_status = 'active'
      AND p.gender_id = ANY($2::int[])
      AND NOT EXISTS (SELECT 1 FROM swipes s
                       WHERE s.actor_id = $1 AND s.target_id = p.user_id)
      AND NOT EXISTS (SELECT 1 FROM swipes s
                       WHERE s.actor_id = p.user_id AND s.target_id = $1
                         AND s.action IN ('like','superlike'))
      AND NOT EXISTS (SELECT 1 FROM user_blocks b
                       WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = p.user_id)
                          OR (b.blocker_user_id = p.user_id AND b.blocked_user_id = $1))
      AND NOT EXISTS (SELECT 1 FROM matches m
                       WHERE (m.user_a = $1 AND m.user_b = p.user_id)
                          OR (m.user_a = p.user_id AND m.user_b = $1))
      AND NOT EXISTS (SELECT 1 FROM daily_picks dp
                       WHERE dp.user_id = $1 AND dp.target_user_id = p.user_id
                         AND dp.status <> 'expired')
),
scored AS (
    SELECT c.user_id AS target_user_id,
        SUM(CASE WHEN uaB.answer_id = ANY(uaA.acceptable_answer_ids)
                 THEN wA.weight ELSE 0 END)::int AS earned_ab,
        SUM(wA.weight)::int                      AS total_ab,
        SUM(CASE WHEN uaA.answer_id = ANY(uaB.acceptable_answer_ids)
                 THEN wB.weight ELSE 0 END)::int AS earned_ba,
        SUM(wB.weight)::int                      AS total_ba,
        COUNT(*)::int                            AS overlap_count,
        BOOL_OR(uaA.importance = 'mandatory'
                AND NOT (uaB.answer_id = ANY(uaA.acceptable_answer_ids))) AS mismatch_ab,
        BOOL_OR(uaB.importance = 'mandatory'
                AND NOT (uaA.answer_id = ANY(uaB.acceptable_answer_ids))) AS mismatch_ba
    FROM candidates c
    JOIN user_answers uaB ON uaB.user_id = c.user_id
    JOIN user_answers uaA ON uaA.user_id = $1
                         AND uaA.question_id = uaB.question_id
    JOIN importance_weights wA ON wA.key = uaA.importance
    JOIN importance_weights wB ON wB.key = uaB.importance
    GROUP BY c.user_id
)
SELECT c.user_id,
    s.overlap_count,
    (  CASE WHEN pr.age_min IS NULL AND pr.age_max IS NULL THEN 1
            WHEN date_part('year', age(c.birthdate))
                 BETWEEN COALESCE(pr.age_min, 0) AND COALESCE(pr.age_max, 200) THEN 1 ELSE 0 END
     + CASE WHEN pr.seek_religion_ids IS NULL OR cardinality(pr.seek_religion_ids) = 0 THEN 1
            WHEN c.religion_id = ANY(pr.seek_religion_ids) THEN 1 ELSE 0 END
     + CASE WHEN pr.seek_sexuality_ids IS NULL OR cardinality(pr.seek_sexuality_ids) = 0 THEN 1
            WHEN c.sexuality_id = ANY(pr.seek_sexuality_ids) THEN 1 ELSE 0 END
     + CASE WHEN pr.seek_ethnicity_ids IS NULL OR cardinality(pr.seek_ethnicity_ids) = 0 THEN 1
            WHEN EXISTS (SELECT 1 FROM user_ethnicities ue
                          WHERE ue.user_id = c.user_id
                            AND ue.ethnicity_id = ANY(pr.seek_ethnicity_ids)) THEN 1 ELSE 0 END
     + CASE WHEN pr.distance_km IS NULL THEN 1
            WHEN ST_Distance(c.geo, v.geo) / 1000.0 <= pr.distance_km THEN 1 ELSE 0 END
    ) AS prefs_satisfied,
    ROUND(100 * sqrt(
          (s.earned_ab::numeric / NULLIF(s.total_ab, 0))
        * (s.earned_ba::numeric / NULLIF(s.total_ba, 0))))::int AS compatibility_percent
FROM candidates c
JOIN scored s ON s.target_user_id = c.user_id
CROSS JOIN prefs pr
CROSS JOIN viewer v
WHERE s.overlap_count >= 5
  AND NOT s.mismatch_ab
  AND NOT s.mismatch_ba
  AND s.total_ab > 0 AND s.total_ba > 0
ORDER BY prefs_satisfied DESC, compatibility_percent DESC,
         md5(c.user_id::text || $1::text)
LIMIT $3;
`

func (r *repository) ScoreCandidates(ctx context.Context, userID string, seekGenderIDs []int16, shortlistSize int) ([]domain.ScoredCandidate, error) {
	rows, err := r.db.QueryContext(ctx, scoreCandidatesQuery, userID, pq.Array(seekGenderIDs), shortlistSize)
	if err != nil {
		return nil, fmt.Errorf("score candidates userID=%s: %w", userID, err)
	}

	defer func() { _ = rows.Close() }()

	var out []domain.ScoredCandidate

	for rows.Next() {
		var c domain.ScoredCandidate
		if err := rows.Scan(&c.UserID, &c.OverlapCount, &c.PrefsSatisfied, &c.CompatibilityPercent); err != nil {
			return nil, fmt.Errorf("scan scored candidate: %w", err)
		}

		out = append(out, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scored candidates: %w", err)
	}

	return out, nil
}

func (r *repository) ReplacePendingBatch(ctx context.Context, tx *sql.Tx, userID string, batchDate time.Time, picks []domain.Pick) (string, error) {
	exec := boil.ContextExecutor(tx)

	if _, err := queries.Raw(
		`DELETE FROM daily_pick_batches WHERE user_id = $1 AND state = 'pending'`,
		userID,
	).ExecContext(ctx, exec); err != nil {
		return "", fmt.Errorf("delete prior pending batch: %w", err)
	}

	var batchID string

	err := queries.Raw(
		`INSERT INTO daily_pick_batches (user_id, batch_date, state)
		 VALUES ($1, $2, 'pending')
		 ON CONFLICT (user_id, batch_date) DO NOTHING
		 RETURNING id`,
		userID, batchDate.Format("2006-01-02"),
	).QueryRowContext(ctx, exec).Scan(&batchID)
	if err == sql.ErrNoRows {
		err = queries.Raw(
			`SELECT id FROM daily_pick_batches WHERE user_id = $1 AND batch_date = $2`,
			userID, batchDate.Format("2006-01-02"),
		).QueryRowContext(ctx, exec).Scan(&batchID)
	}

	if err != nil {
		return "", fmt.Errorf("insert pending batch: %w", err)
	}

	if err := r.insertPicks(ctx, exec, batchID, userID, picks); err != nil {
		return "", err
	}

	return batchID, nil
}

func (r *repository) InsertPicks(ctx context.Context, tx *sql.Tx, batchID, userID string, picks []domain.Pick) error {
	return r.insertPicks(ctx, tx, batchID, userID, picks)
}

func (r *repository) insertPicks(ctx context.Context, exec boil.ContextExecutor, batchID, userID string, picks []domain.Pick) error {
	for _, p := range picks {
		_, err := queries.Raw(
			`INSERT INTO daily_picks (batch_id, user_id, target_user_id, position,
			                          compatibility_percent, overlap_count, prefs_satisfied, status)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`,
			batchID, userID, p.TargetUserID, p.Position,
			p.CompatibilityPercent, p.OverlapCount, p.PrefsSatisfied,
		).ExecContext(ctx, exec)
		if err != nil {
			return fmt.Errorf("insert pick target=%s: %w", p.TargetUserID, err)
		}
	}

	return nil
}

func (r *repository) execOrDB(exec boil.ContextExecutor) boil.ContextExecutor {
	if exec == nil {
		return r.db
	}

	return exec
}

func (r *repository) GetBatchByState(ctx context.Context, exec boil.ContextExecutor, userID, state string) (*domain.Batch, error) {
	exec = r.execOrDB(exec)

	const q = `
		SELECT id, user_id, batch_date, state, revealed_at, notified_at
		FROM daily_pick_batches
		WHERE user_id = $1 AND state = $2
		LIMIT 1
	`

	var b domain.Batch

	var revealedAt, notifiedAt sql.NullTime

	err := queries.Raw(q, userID, state).QueryRowContext(ctx, exec).Scan(
		&b.ID, &b.UserID, &b.BatchDate, &b.State, &revealedAt, &notifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get batch by state: %w", err)
	}

	if revealedAt.Valid {
		t := revealedAt.Time
		b.RevealedAt = &t
	}

	if notifiedAt.Valid {
		t := notifiedAt.Time
		b.NotifiedAt = &t
	}

	picks, err := r.loadPicksForBatch(ctx, exec, b.ID)
	if err != nil {
		return nil, err
	}

	b.Picks = picks

	return &b, nil
}

func (r *repository) GetBatchByID(ctx context.Context, exec boil.ContextExecutor, batchID string) (*domain.Batch, error) {
	exec = r.execOrDB(exec)

	const q = `
		SELECT id, user_id, batch_date, state, revealed_at, notified_at
		FROM daily_pick_batches
		WHERE id = $1
	`

	var b domain.Batch

	var revealedAt, notifiedAt sql.NullTime

	err := queries.Raw(q, batchID).QueryRowContext(ctx, exec).Scan(
		&b.ID, &b.UserID, &b.BatchDate, &b.State, &revealedAt, &notifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get batch by id: %w", err)
	}

	if revealedAt.Valid {
		t := revealedAt.Time
		b.RevealedAt = &t
	}

	if notifiedAt.Valid {
		t := notifiedAt.Time
		b.NotifiedAt = &t
	}

	picks, err := r.loadPicksForBatch(ctx, exec, b.ID)
	if err != nil {
		return nil, err
	}

	b.Picks = picks

	return &b, nil
}

func (r *repository) loadPicksForBatch(ctx context.Context, exec boil.ContextExecutor, batchID string) ([]domain.Pick, error) {
	const q = `
		SELECT id, batch_id, user_id, target_user_id, position,
		       compatibility_percent, overlap_count, prefs_satisfied, status, decided_at
		FROM daily_picks
		WHERE batch_id = $1
		ORDER BY position ASC
	`

	rows, err := queries.Raw(q, batchID).QueryContext(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("load picks: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var picks []domain.Pick

	for rows.Next() {
		var p domain.Pick

		var decidedAt sql.NullTime
		if err := rows.Scan(
			&p.ID, &p.BatchID, &p.UserID, &p.TargetUserID, &p.Position,
			&p.CompatibilityPercent, &p.OverlapCount, &p.PrefsSatisfied, &p.Status, &decidedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pick: %w", err)
		}

		if decidedAt.Valid {
			t := decidedAt.Time
			p.DecidedAt = &t
		}

		picks = append(picks, p)
	}

	return picks, rows.Err()
}

func (r *repository) CountUndecidedPicks(ctx context.Context, exec boil.ContextExecutor, batchID string) (int, error) {
	exec = r.execOrDB(exec)

	const q = `SELECT COUNT(*) FROM daily_picks WHERE batch_id = $1 AND status = 'pending'`

	var count int
	if err := queries.Raw(q, batchID).QueryRowContext(ctx, exec).Scan(&count); err != nil {
		return 0, fmt.Errorf("count undecided picks: %w", err)
	}

	return count, nil
}

func (r *repository) MarkBatchState(ctx context.Context, tx *sql.Tx, batchID, state string, revealedAt *time.Time) error {
	exec := boil.ContextExecutor(tx)

	_, err := queries.Raw(
		`UPDATE daily_pick_batches
		 SET state = $2, revealed_at = COALESCE($3, revealed_at), updated_at = now()
		 WHERE id = $1`,
		batchID, state, revealedAt,
	).ExecContext(ctx, exec)
	if err != nil {
		return fmt.Errorf("mark batch state: %w", err)
	}

	return nil
}

func (r *repository) MarkPickDecided(ctx context.Context, exec boil.ContextExecutor, userID, targetUserID, status string) error {
	if exec == nil {
		exec = r.db
	}

	res, err := queries.Raw(
		`UPDATE daily_picks
		 SET status = $3, decided_at = now(), updated_at = now()
		 WHERE user_id = $1 AND target_user_id = $2 AND status = 'pending'`,
		userID, targetUserID, status,
	).ExecContext(ctx, exec)
	if err != nil {
		return fmt.Errorf("mark pick decided: %w", err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return nil
	}

	return nil
}

func (r *repository) ExpirePicks(ctx context.Context, tx *sql.Tx, pickIDs []string) error {
	if len(pickIDs) == 0 {
		return nil
	}

	_, err := queries.Raw(
		`UPDATE daily_picks SET status = 'expired', updated_at = now() WHERE id = ANY($1)`,
		pq.Array(pickIDs),
	).ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("expire picks: %w", err)
	}

	return nil
}

func (r *repository) ListEligibleUserIDs(ctx context.Context) ([]string, error) {
	const q = `
		SELECT id FROM users
		WHERE onboarding_step = 'COMPLETE' AND account_status = 'active'
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list eligible user IDs: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (r *repository) LockUserForPicks(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("lock user for picks userID=%s: %w", userID, err)
	}

	return nil
}

func (r *repository) SeekGenderIDs(ctx context.Context, userID string) ([]int16, error) {
	const q = `SELECT seek_gender_ids FROM user_preferences WHERE user_id = $1`

	var ids pq.Int64Array

	err := r.db.QueryRowContext(ctx, q, userID).Scan(&ids)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("seek gender ids: %w", err)
	}

	out := make([]int16, len(ids))
	for i, id := range ids {
		out[i] = int16(id)
	}

	return out, nil
}

func (r *repository) CountActiveMatches(ctx context.Context, exec boil.ContextExecutor, userID string) (int64, error) {
	exec = r.execOrDB(exec)

	const q = `
		SELECT COUNT(*) FROM matches
		WHERE (user_a = $1 OR user_b = $1) AND status = 'active'
	`

	var total int64
	if err := queries.Raw(q, userID).QueryRowContext(ctx, exec).Scan(&total); err != nil {
		return 0, fmt.Errorf("count active matches userID=%s: %w", userID, err)
	}

	return total, nil
}

func (r *repository) HasAnyBatch(ctx context.Context, exec boil.ContextExecutor, userID string) (bool, error) {
	exec = r.execOrDB(exec)

	const q = `SELECT EXISTS (SELECT 1 FROM daily_pick_batches WHERE user_id = $1)`

	var exists bool
	if err := queries.Raw(q, userID).QueryRowContext(ctx, exec).Scan(&exists); err != nil {
		return false, fmt.Errorf("has any batch: %w", err)
	}

	return exists, nil
}

func (r *repository) IsTargetStillValid(ctx context.Context, viewerID, targetID string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM user_profiles p
			JOIN users u ON u.id = p.user_id
			WHERE p.user_id = $2
			  AND u.onboarding_step = 'COMPLETE'
			  AND u.account_status = 'active'
		)
		AND NOT EXISTS (
			SELECT 1 FROM user_blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = $2)
			   OR (b.blocker_user_id = $2 AND b.blocked_user_id = $1)
		)
		AND NOT EXISTS (
			SELECT 1 FROM matches m
			WHERE (m.user_a = $1 AND m.user_b = $2)
			   OR (m.user_a = $2 AND m.user_b = $1)
		)
	`

	var valid bool
	if err := r.db.QueryRowContext(ctx, q, viewerID, targetID).Scan(&valid); err != nil {
		return false, fmt.Errorf("is target still valid: %w", err)
	}

	return valid, nil
}

func (r *repository) MarkBatchNotified(ctx context.Context, batchIDs []string) error {
	if len(batchIDs) == 0 {
		return nil
	}

	_, err := queries.Raw(
		`UPDATE daily_pick_batches SET notified_at = now(), updated_at = now() WHERE id = ANY($1)`,
		pq.Array(batchIDs),
	).ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("mark batch notified: %w", err)
	}

	return nil
}

func (r *repository) ListPendingPickTargetIDs(ctx context.Context, batchID string) ([]string, error) {
	const q = `
		SELECT target_user_id FROM daily_picks
		WHERE batch_id = $1 AND status = 'pending'
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, q, batchID)
	if err != nil {
		return nil, fmt.Errorf("list pending pick targets: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}
