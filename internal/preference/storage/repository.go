package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type PreferenceRepository interface {
	InsertPreference(ctx context.Context, tx *sql.Tx, preference *entity.UserPreference) error
	IsAnalyticsOptedOut(ctx context.Context, userID string) (bool, error)
	SetAnalyticsOptOut(ctx context.Context, userID string, optedOut bool) error
}

type preferenceRepository struct {
	db *sqlx.DB
}

func NewPreferenceRepository(db *sqlx.DB) PreferenceRepository {
	return &preferenceRepository{
		db: db,
	}
}

func (pr *preferenceRepository) InsertPreference(ctx context.Context, tx *sql.Tx, preference *entity.UserPreference) error {
	err := preference.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (pr *preferenceRepository) IsAnalyticsOptedOut(ctx context.Context, userID string) (bool, error) {
	var optedOut bool

	err := pr.db.QueryRowContext(ctx,
		`SELECT analytics_opt_out FROM user_preferences WHERE user_id = $1`, userID,
	).Scan(&optedOut)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return optedOut, nil
}

func (pr *preferenceRepository) SetAnalyticsOptOut(ctx context.Context, userID string, optedOut bool) error {
	res, err := pr.db.ExecContext(ctx,
		`UPDATE user_preferences SET analytics_opt_out = $2 WHERE user_id = $1`, userID, optedOut,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
