package storage

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface {
	GetDatingIntentions(ctx context.Context) (entity.DatingIntentionSlice, error)
	GetGenders(ctx context.Context) (entity.GenderSlice, error)
}

type onboardingRepository struct {
	db *sqlx.DB
}

func NewOnboardingRepository(db *sqlx.DB) OnboardingRepository {
	return &onboardingRepository{
		db: db,
	}
}

func (or *onboardingRepository) GetDatingIntentions(ctx context.Context) (entity.DatingIntentionSlice, error) {
	di, err := entity.DatingIntentions().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return di, nil
}

func (or *onboardingRepository) GetGenders(ctx context.Context) (entity.GenderSlice, error) {
	genders, err := entity.Genders().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return genders, nil
}
