package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type Repository interface {
	CountMessages(ctx context.Context, from, to time.Time) (messages int64, voice int64, err error)
	TopPromptWeekly(ctx context.Context, weekStart time.Time) (promptID *int64, likeRate *float64, err error)
	UserCounts(ctx context.Context, userID string, from, to time.Time) (likesSent, likesReceived, matches, messages, voice int64, err error)
	InsertGlobalWeeklySnapshot(ctx context.Context, key string, from, to time.Time, payload any) error
	UpsertWrapped(ctx context.Context, userID string, year int, payload any) error
	GetAppOpenEvents(ctx context.Context, userID string, from, to time.Time) ([]*entity.Event, error)
	GetLatestAppOpen(ctx context.Context, userID string) (*entity.Event, error)
	GetFirstAppOpen(ctx context.Context, userID string) (*entity.Event, error)
	GetDailyActiveUsers(ctx context.Context, date time.Time) (int64, error)
	GetWeeklyActiveUsers(ctx context.Context, weekStart time.Time) (int64, error)
	GetMonthlyActiveUsers(ctx context.Context, monthStart time.Time) (int64, error)
	GetRetentionCohort(ctx context.Context, signupDate time.Time, daysAfter int) (cohortSize int64, returned int64, err error)
	GetAverageTimeToFirstReturn(ctx context.Context, from, to time.Time) (*time.Duration, error)
	GetAverageTimeBetweenOpens(ctx context.Context, userID string) (*time.Duration, error)
	GetSessionFrequency(ctx context.Context, userID string, from, to time.Time) (int64, error)
	GetUserSignupDate(ctx context.Context, userID string) (time.Time, error)
	GetTotalUsers(ctx context.Context) (int64, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CountMessages(ctx context.Context, from, to time.Time) (int64, int64, error) {
	total, err := entity.Messages(
		qm.Where(entity.MessageTableColumns.CreatedAt+" >= ? AND "+entity.MessageTableColumns.CreatedAt+" < ?", from, to),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, err
	}

	voice, err := entity.Messages(
		qm.Where(entity.MessageTableColumns.CreatedAt+" >= ? AND "+entity.MessageTableColumns.CreatedAt+" < ?", from, to),
		qm.Where(entity.MessageTableColumns.Type+" = ?", entity.MessageTypeVoice),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, err
	}

	return total, voice, nil
}

func (r *repository) TopPromptWeekly(ctx context.Context, weekStart time.Time) (*int64, *float64, error) {
	// Pull rows for that week and compute like rate in Go, then pick best.
	rows, err := entity.VPromptPopularities(
		qm.Where(entity.VPromptPopularityTableColumns.WeekStart+" = ?", weekStart),
	).All(ctx, r.db)
	if err != nil {
		return nil, nil, err
	}

	var bestID *int64

	var bestRate *float64

	for _, row := range rows {
		if row == nil || row.PromptID.IsZero() || row.TotalSwipes.IsZero() {
			continue
		}

		likes := row.Likes.Int64

		total := row.TotalSwipes.Int64
		if total == 0 {
			continue
		}

		rate := float64(likes) / float64(total)
		if bestRate == nil || rate > *bestRate {
			id := row.PromptID.Int64
			bestID = &id
			br := rate
			bestRate = &br
		}
	}

	return bestID, bestRate, nil
}

func (r *repository) UserCounts(ctx context.Context, userID string, from, to time.Time) (int64, int64, int64, int64, int64, error) {
	likesSent, err := entity.Swipes(
		qm.Where(entity.SwipeTableColumns.ActorID+" = ?", userID),
		qm.Where(entity.SwipeTableColumns.CreatedAt+" >= ? AND "+entity.SwipeTableColumns.CreatedAt+" < ?", from, to),
		qm.Where(entity.SwipeTableColumns.Action+" IN ('like','superlike')"),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	likesReceived, err := entity.Swipes(
		qm.Where(entity.SwipeTableColumns.TargetID+" = ?", userID),
		qm.Where(entity.SwipeTableColumns.CreatedAt+" >= ? AND "+entity.SwipeTableColumns.CreatedAt+" < ?", from, to),
		qm.Where(entity.SwipeTableColumns.Action+" IN ('like','superlike')"),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	matches, err := entity.Matches(
		qm.Where("("+entity.MatchTableColumns.UserA+" = ? OR "+entity.MatchTableColumns.UserB+" = ?)", userID, userID),
		qm.Where(entity.MatchTableColumns.CreatedAt+" >= ? AND "+entity.MatchTableColumns.CreatedAt+" < ?", from, to),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	messages, err := entity.Messages(
		qm.Where(entity.MessageTableColumns.SenderID+" = ?", userID),
		qm.Where(entity.MessageTableColumns.CreatedAt+" >= ? AND "+entity.MessageTableColumns.CreatedAt+" < ?", from, to),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	voice, err := entity.Messages(
		qm.Where(entity.MessageTableColumns.SenderID+" = ?", userID),
		qm.Where(entity.MessageTableColumns.CreatedAt+" >= ? AND "+entity.MessageTableColumns.CreatedAt+" < ?", from, to),
		qm.Where(entity.MessageTableColumns.Type+" = ?", entity.MessageTypeVoice),
	).Count(ctx, r.db)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return likesSent, likesReceived, matches, messages, voice, nil
}

func (r *repository) InsertGlobalWeeklySnapshot(ctx context.Context, key string, from, to time.Time, payload any) error {
	b, _ := json.Marshal(payload)
	rec := &entity.InsightSnapshot{
		ID:          "",
		Key:         key,
		PeriodStart: from,
		PeriodEnd:   to,
		Scope:       "global",
		Payload:     types.JSON(b),
	}

	return rec.Insert(ctx, r.db, boil.Infer())
}

func (r *repository) UpsertWrapped(ctx context.Context, userID string, year int, payload any) error {
	b, _ := json.Marshal(payload)
	rec := &entity.WrappedAnnual{
		UserID:    userID,
		Year:      year,
		Payload:   types.JSON(b),
		CreatedAt: time.Now().UTC(),
	}
	// Use Upsert to handle conflict
	return rec.Upsert(ctx, r.db, true,
		[]string{entity.WrappedAnnualColumns.UserID, entity.WrappedAnnualColumns.Year},
		boil.Whitelist(entity.WrappedAnnualColumns.Payload),
		boil.Infer(),
	)
}

// GetAppOpenEvents retrieves app open events for a user in a time range
func (r *repository) GetAppOpenEvents(ctx context.Context, userID string, from, to time.Time) ([]*entity.Event, error) {
	events, err := entity.Events(
		qm.Where(entity.EventTableColumns.Name+" = ?", "app.opened"),
		qm.Where(entity.EventTableColumns.UserID+" = ?", userID),
		qm.Where(entity.EventTableColumns.OccurredAt+" >= ? AND "+entity.EventTableColumns.OccurredAt+" < ?", from, to),
		qm.OrderBy(entity.EventTableColumns.OccurredAt+" ASC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetLatestAppOpen retrieves the most recent app open for a user
func (r *repository) GetLatestAppOpen(ctx context.Context, userID string) (*entity.Event, error) {
	event, err := entity.Events(
		qm.Where(entity.EventTableColumns.Name+" = ?", "app.opened"),
		qm.Where(entity.EventTableColumns.UserID+" = ?", userID),
		qm.OrderBy(entity.EventTableColumns.OccurredAt+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetFirstAppOpen retrieves the first app open for a user
func (r *repository) GetFirstAppOpen(ctx context.Context, userID string) (*entity.Event, error) {
	event, err := entity.Events(
		qm.Where(entity.EventTableColumns.Name+" = ?", "app.opened"),
		qm.Where(entity.EventTableColumns.UserID+" = ?", userID),
		qm.OrderBy(entity.EventTableColumns.OccurredAt+" ASC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetDailyActiveUsers counts unique users who opened app on a specific date
func (r *repository) GetDailyActiveUsers(ctx context.Context, date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	var distinctCount int64

	err := r.db.GetContext(ctx, &distinctCount,
		`SELECT COUNT(DISTINCT user_id) FROM events WHERE name = $1 AND occurred_at >= $2 AND occurred_at < $3`,
		"app.opened", start, end)
	if err != nil {
		return 0, err
	}

	return distinctCount, nil
}

// GetWeeklyActiveUsers counts unique users who opened app in a week
func (r *repository) GetWeeklyActiveUsers(ctx context.Context, weekStart time.Time) (int64, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)

	var distinctCount int64

	err := r.db.GetContext(ctx, &distinctCount,
		`SELECT COUNT(DISTINCT user_id) FROM events WHERE name = $1 AND occurred_at >= $2 AND occurred_at < $3`,
		"app.opened", weekStart, weekEnd)
	if err != nil {
		return 0, err
	}

	return distinctCount, nil
}

// GetMonthlyActiveUsers counts unique users who opened app in a month
func (r *repository) GetMonthlyActiveUsers(ctx context.Context, monthStart time.Time) (int64, error) {
	monthEnd := monthStart.AddDate(0, 1, 0)

	var distinctCount int64

	err := r.db.GetContext(ctx, &distinctCount,
		`SELECT COUNT(DISTINCT user_id) FROM events WHERE name = $1 AND occurred_at >= $2 AND occurred_at < $3`,
		"app.opened", monthStart, monthEnd)
	if err != nil {
		return 0, err
	}

	return distinctCount, nil
}

// GetRetentionCohort gets retention for users who signed up on a specific date, checked N days later
func (r *repository) GetRetentionCohort(ctx context.Context, signupDate time.Time, daysAfter int) (int64, int64, error) {
	cohortStart := time.Date(signupDate.Year(), signupDate.Month(), signupDate.Day(), 0, 0, 0, 0, time.UTC)
	cohortEnd := cohortStart.AddDate(0, 0, 1)
	checkStart := cohortStart.AddDate(0, 0, daysAfter)
	checkEnd := checkStart.AddDate(0, 0, 1)

	// Count users who signed up in the cohort period
	var cohortSize int64

	err := r.db.GetContext(ctx, &cohortSize,
		`SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2`,
		cohortStart, cohortEnd)
	if err != nil {
		return 0, 0, err
	}

	// Count users from the cohort who opened the app during the check period
	var returned int64

	err = r.db.GetContext(ctx, &returned,
		`SELECT COUNT(DISTINCT e.user_id) 
		 FROM events e
		 INNER JOIN users u ON e.user_id = u.id
		 WHERE e.name = $1 
		   AND u.created_at >= $2 AND u.created_at < $3
		   AND e.occurred_at >= $4 AND e.occurred_at < $5`,
		"app.opened", cohortStart, cohortEnd, checkStart, checkEnd)
	if err != nil {
		return 0, 0, err
	}

	return cohortSize, returned, nil
}

// GetAverageTimeToFirstReturn calculates average time between signup and first app open
func (r *repository) GetAverageTimeToFirstReturn(ctx context.Context, from, to time.Time) (*time.Duration, error) {
	var avgSeconds *float64

	err := r.db.GetContext(ctx, &avgSeconds,
		`SELECT AVG(EXTRACT(EPOCH FROM (e.occurred_at - u.created_at))) as avg_seconds
		 FROM events e
		 INNER JOIN users u ON e.user_id = u.id
		 WHERE e.name = $1
		   AND u.created_at >= $2 AND u.created_at < $3
		   AND e.occurred_at = (
		     SELECT MIN(e2.occurred_at)
		     FROM events e2
		     WHERE e2.user_id = e.user_id AND e2.name = $1
		   )`,
		"app.opened", from, to)
	if err != nil {
		return nil, err
	}

	if avgSeconds == nil {
		return nil, nil
	}

	duration := time.Duration(*avgSeconds) * time.Second

	return &duration, nil
}

// GetAverageTimeBetweenOpens calculates average time between consecutive app opens for a user
func (r *repository) GetAverageTimeBetweenOpens(ctx context.Context, userID string) (*time.Duration, error) {
	var avgSeconds *float64

	err := r.db.GetContext(ctx, &avgSeconds,
		`SELECT AVG(EXTRACT(EPOCH FROM (curr.occurred_at - prev.occurred_at))) as avg_seconds
		 FROM events curr
		 INNER JOIN events prev ON curr.user_id = prev.user_id
		   AND prev.occurred_at = (
		     SELECT MAX(e.occurred_at)
		     FROM events e
		     WHERE e.user_id = curr.user_id
		       AND e.name = $1
		       AND e.occurred_at < curr.occurred_at
		   )
		 WHERE curr.user_id = $2
		   AND curr.name = $1
		   AND prev.name = $1`,
		"app.opened", userID)
	if err != nil {
		return nil, err
	}

	if avgSeconds == nil {
		return nil, nil
	}

	duration := time.Duration(*avgSeconds) * time.Second

	return &duration, nil
}

// GetSessionFrequency counts app opens (sessions) for a user in a time range
func (r *repository) GetSessionFrequency(ctx context.Context, userID string, from, to time.Time) (int64, error) {
	count, err := entity.Events(
		qm.Where(entity.EventTableColumns.Name+" = ?", "app.opened"),
		qm.Where(entity.EventTableColumns.UserID+" = ?", userID),
		qm.Where(entity.EventTableColumns.OccurredAt+" >= ? AND "+entity.EventTableColumns.OccurredAt+" < ?", from, to),
	).Count(ctx, r.db)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetUserSignupDate retrieves the signup date (created_at) for a user
func (r *repository) GetUserSignupDate(ctx context.Context, userID string) (time.Time, error) {
	user, err := entity.Users(entity.UserWhere.ID.EQ(userID)).One(ctx, r.db)
	if err != nil {
		return time.Time{}, err
	}

	return user.CreatedAt, nil
}

// GetTotalUsers counts total users in the system
func (r *repository) GetTotalUsers(ctx context.Context) (int64, error) {
	count, err := entity.Users().Count(ctx, r.db)
	if err != nil {
		return 0, err
	}

	return count, nil
}
