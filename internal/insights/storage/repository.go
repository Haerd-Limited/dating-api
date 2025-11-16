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
