package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

var ErrModerationWarningNotFound = errors.New("moderation warning not found")

func (r *repository) InsertModerationWarning(ctx context.Context, warning *entity.UserModerationWarning, tx *sql.Tx) error {
	exec := r.executor(tx)
	return warning.Insert(ctx, exec, boil.Infer())
}

func (r *repository) ListUnacknowledgedWarnings(ctx context.Context, userID string) (entity.UserModerationWarningSlice, error) {
	return entity.UserModerationWarnings(
		entity.UserModerationWarningWhere.UserID.EQ(userID),
		entity.UserModerationWarningWhere.AcknowledgedAt.IsNull(),
		qm.OrderBy(entity.UserModerationWarningColumns.CreatedAt+" DESC"),
	).All(ctx, r.db)
}

func (r *repository) GetWarningByID(ctx context.Context, warningID string) (*entity.UserModerationWarning, error) {
	warning, err := entity.UserModerationWarnings(
		entity.UserModerationWarningWhere.ID.EQ(warningID),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get warning by id: %w", err)
	}

	return warning, nil
}

func (r *repository) AcknowledgeWarning(ctx context.Context, warningID, userID string) error {
	warning, err := r.GetWarningByID(ctx, warningID)
	if err != nil {
		return err
	}

	if warning == nil || warning.UserID != userID {
		return ErrModerationWarningNotFound
	}

	if warning.AcknowledgedAt.Valid {
		return nil
	}

	warning.AcknowledgedAt = null.TimeFrom(time.Now().UTC())
	_, err = warning.Update(ctx, r.db, boil.Whitelist(entity.UserModerationWarningColumns.AcknowledgedAt))

	return err
}
