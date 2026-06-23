package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

type accountGateRow struct {
	AccountStatus     string      `db:"account_status"`
	SuspendedUntil    null.Time   `db:"suspended_until"`
	ModerationReason  null.String `db:"moderation_reason"`
	HasPendingWarning bool        `db:"has_pending_warning"`
}

func (r *userRepository) GetAccountGateState(ctx context.Context, userID string) (*domain.AccountState, error) {
	var row accountGateRow

	err := r.db.GetContext(ctx, &row, `
		SELECT
			u.account_status,
			u.suspended_until,
			u.moderation_reason,
			EXISTS (
				SELECT 1 FROM user_moderation_warnings w
				WHERE w.user_id = u.id AND w.acknowledged_at IS NULL
			) AS has_pending_warning
		FROM users u
		WHERE u.id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserDoesNotExists
		}

		return nil, fmt.Errorf("get account gate state userID=%s: %w", userID, err)
	}

	state := &domain.AccountState{
		Status:            row.AccountStatus,
		HasPendingWarning: row.HasPendingWarning,
	}

	if row.SuspendedUntil.Valid {
		t := row.SuspendedUntil.Time
		state.SuspendedUntil = &t
	}

	if row.ModerationReason.Valid {
		reason := row.ModerationReason.String
		state.Reason = &reason
	}

	return state, nil
}

func (r *userRepository) UpdateAccountStatus(
	ctx context.Context,
	userID, status string,
	suspendedUntil *time.Time,
	reason *string,
	tx *sql.Tx,
) error {
	exec := boilExecutor(r.db, tx)
	now := time.Now().UTC()

	var suspendedUntilVal any
	if suspendedUntil != nil {
		suspendedUntilVal = *suspendedUntil
	}

	var reasonVal any
	if reason != nil {
		reasonVal = *reason
	}

	_, err := exec.ExecContext(ctx, `
		UPDATE users
		SET account_status = $1,
		    suspended_until = $2,
		    moderation_reason = $3,
		    status_updated_at = $4,
		    updated_at = $4
		WHERE id = $5
	`, status, suspendedUntilVal, reasonVal, now, userID)
	if err != nil {
		return fmt.Errorf("update account status userID=%s: %w", userID, err)
	}

	return nil
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func boilExecutor(db *sqlx.DB, tx *sql.Tx) execContext {
	if tx != nil {
		return tx
	}

	return db
}
