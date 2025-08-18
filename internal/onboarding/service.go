package onboarding

import (
	"context"
	"fmt"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	onboardingstorage "github.com/Haerd-Limited/dating-api/internal/onboarding/storage"
)

type Service interface {
	Patch(ctx context.Context, update domain.OnboardingUpdate) error
}

type onboardingService struct {
	logger *zap.Logger
	repo   onboardingstorage.OnboardingRepository
}

func NewOnboardingService(
	logger *zap.Logger,
	onboardingRepository onboardingstorage.OnboardingRepository,
) Service {
	return &onboardingService{
		logger: logger,
		repo:   onboardingRepository,
	}
}

var ErrInvalidAgePreference = errors.New("min age cannot be greater than max age")

func (os *onboardingService) Patch(ctx context.Context, update domain.OnboardingUpdate) error {
	// ---- Basic validation ----
	if update.Preferences != nil &&
		update.Preferences.AgeMin != nil &&
		update.Preferences.AgeMax != nil &&
		*update.Preferences.AgeMin > *update.Preferences.AgeMax {
		return ErrInvalidAgePreference
	}

	// ---- Map DOMAIN -> ENTITY ----
	profileEnt, err := mapper.MapProfileToEntity(update.UserID, update.UserProfile)
	if err != nil {
		return fmt.Errorf("failed to map profile: %w", err)
	}

	preferencesEnt, err := mapper.MapPreferencesToEntity(update.UserID, update.Preferences)
	if err != nil {
		return fmt.Errorf("failed to map preferences: %w", err)
	}

	// NOTE:
	// - language/interests slices are passed through as-is; the repo will replace the M2M rows when provided.
	// - geo (lat/lon) isn’t set here because the geography type isn’t handled by sqlboiler directly.
	//   Your repo already owns DB access; if you want to persist geo, extend the repo to take lat/lon
	//   and issue an `UPDATE ... SET geo = ST_SetSRID(ST_MakePoint(lon,lat),4326)::geography`.

	bump := false
	if update.BumpOnboardingStep != nil {
		bump = *update.BumpOnboardingStep
	}

	// ✅ Persist update
	return os.repo.PatchOnboardingTx(
		ctx,
		update.UserID,
		profileEnt,
		preferencesEnt,
		update.LanguageIDs,
		update.InterestIDs,
		bump,
		update.UserProfile.Latitude,
		update.UserProfile.Longitude,
	)
}
