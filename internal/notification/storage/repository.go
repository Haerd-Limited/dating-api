package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type DeviceTokenRepository interface {
	UpsertToken(ctx context.Context, userID, token string) error
	DeleteToken(ctx context.Context, userID, token string) error
	DeleteTokens(ctx context.Context, tokens []string) error
	GetTokensByUserIDs(ctx context.Context, userIDs []string) (map[string][]string, error)
	ListUserIDsWithTokens(ctx context.Context) ([]string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewDeviceTokenRepository(db *sqlx.DB) DeviceTokenRepository {
	return &repository{db: db}
}

func (r *repository) UpsertToken(ctx context.Context, userID, token string) error {
	const query = `
INSERT INTO device_tokens (user_id, token)
VALUES ($1, $2)
ON CONFLICT (user_id, token) DO UPDATE
SET updated_at = now()`

	if _, err := r.db.ExecContext(ctx, query, userID, token); err != nil {
		return fmt.Errorf("upsert device token userID=%s: %w", userID, err)
	}

	return nil
}

func (r *repository) DeleteToken(ctx context.Context, userID, token string) error {
	const query = `DELETE FROM device_tokens WHERE user_id = $1 AND token = $2`

	res, err := r.db.ExecContext(ctx, query, userID, token)
	if err != nil {
		return fmt.Errorf("delete device token userID=%s: %w", userID, err)
	}

	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *repository) DeleteTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}

	const query = `DELETE FROM device_tokens WHERE token = ANY($1)`

	if _, err := r.db.ExecContext(ctx, query, pq.StringArray(tokens)); err != nil {
		return fmt.Errorf("delete device tokens: %w", err)
	}

	return nil
}

func (r *repository) GetTokensByUserIDs(ctx context.Context, userIDs []string) (result map[string][]string, err error) {
	if len(userIDs) == 0 {
		return map[string][]string{}, nil
	}

	const query = `
SELECT user_id, token
FROM device_tokens
WHERE user_id = ANY($1)`

	rows, err := r.db.QueryxContext(ctx, query, pq.StringArray(userIDs))
	if err != nil {
		return nil, fmt.Errorf("query device tokens: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close device token rows: %w", closeErr)
			result = nil
		}
	}()

	result = make(map[string][]string, len(userIDs))

	for rows.Next() {
		var row struct {
			UserID string `db:"user_id"`
			Token  string `db:"token"`
		}

		if err = rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("scan device token row: %w", err)
		}

		result[row.UserID] = append(result[row.UserID], row.Token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device token rows: %w", err)
	}

	return result, nil
}

func (r *repository) ListUserIDsWithTokens(ctx context.Context) (userIDs []string, err error) {
	const query = `SELECT DISTINCT user_id FROM device_tokens`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list device token users: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close device token user rows: %w", closeErr)
			userIDs = nil
		}
	}()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan device token user: %w", err)
		}

		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device token users: %w", err)
	}

	return userIDs, nil
}
