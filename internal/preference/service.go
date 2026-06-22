//go:generate mockgen -source=service.go -destination=service_mock.go -package=preference
package preference

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/preference/domain"
	"github.com/Haerd-Limited/dating-api/internal/preference/storage"
)

type Service interface {
	ScaffoldUserPreferences(ctx context.Context, tx *sql.Tx, userID string) error
	IsAnalyticsOptedOut(ctx context.Context, userID string) (bool, error)
	SetAnalyticsOptOut(ctx context.Context, userID string, optedOut bool) error
}

type service struct {
	logger         *zap.Logger
	preferenceRepo storage.PreferenceRepository
}

func NewPreferenceService(
	logger *zap.Logger,
	preferenceRepo storage.PreferenceRepository,
) Service {
	return &service{
		logger:         logger,
		preferenceRepo: preferenceRepo,
	}
}

func (s *service) ScaffoldUserPreferences(ctx context.Context, tx *sql.Tx, userID string) error {
	err := s.preferenceRepo.InsertPreference(ctx, tx, &entity.UserPreference{
		UserID:     userID,
		DistanceKM: null.Int16From(domain.DefaultDistanceKM),
		AgeMin:     null.Int16From(domain.MinAge),
		AgeMax:     null.Int16From(domain.MaxAge),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) IsAnalyticsOptedOut(ctx context.Context, userID string) (bool, error) {
	optedOut, err := s.preferenceRepo.IsAnalyticsOptedOut(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("is analytics opted out: %w", err)
	}

	return optedOut, nil
}

func (s *service) SetAnalyticsOptOut(ctx context.Context, userID string, optedOut bool) error {
	err := s.preferenceRepo.SetAnalyticsOptOut(ctx, userID, optedOut)
	if err != nil {
		return fmt.Errorf("set analytics opt out: %w", err)
	}

	return nil
}
