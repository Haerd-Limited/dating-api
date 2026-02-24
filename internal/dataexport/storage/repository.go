package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository provides rate limiting and audit for data export requests.
type Repository interface {
	GetLastRequestedAt(ctx context.Context, userID string) (time.Time, error)
	InsertRequest(ctx context.Context, userID string) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository returns a new data export storage repository.
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// GetLastRequestedAt returns the most recent requested_at for the user, or zero time if none.
func (r *repository) GetLastRequestedAt(ctx context.Context, userID string) (time.Time, error) {
	const query = `SELECT requested_at FROM data_export_requests WHERE user_id = $1 ORDER BY requested_at DESC LIMIT 1`
	var t time.Time
	err := r.db.GetContext(ctx, &t, query, userID)
	if err != nil {
		if isNoRows(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return t, nil
}

// InsertRequest records a data export request for audit and rate limiting.
func (r *repository) InsertRequest(ctx context.Context, userID string) error {
	const query = `INSERT INTO data_export_requests (user_id) VALUES ($1)`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func isNoRows(err error) bool {
	return err != nil && errors.Is(err, sql.ErrNoRows)
}
