package storage

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type PreferenceRepository interface {
	InsertPreference(ctx context.Context, tx *sql.Tx, preference *entity.UserPreference) error
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
