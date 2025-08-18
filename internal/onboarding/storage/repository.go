package storage

import (
	"github.com/jmoiron/sqlx"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface{}

type onboardingRepository struct {
	db *sqlx.DB
}

func NewOnboardingRepository(db *sqlx.DB) OnboardingRepository {
	return &onboardingRepository{
		db: db,
	}
}
