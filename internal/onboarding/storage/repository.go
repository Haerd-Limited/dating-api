package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface {
	PatchOnboardingTx(
		ctx context.Context,
		userID string,
		profile *entity.UserProfile,
		prefs *entity.UserPreference,
		languageIDs *[]int32,
		interestIDs *[]int32,
		bumpStep bool,
		latitude *float64,
		longitude *float64,
	) error
}

type onboardingRepository struct {
	db *sqlx.DB
}

func NewOnboardingRepository(db *sqlx.DB) OnboardingRepository {
	return &onboardingRepository{
		db: db,
	}
}

func (r *onboardingRepository) PatchOnboardingTx(
	ctx context.Context,
	userID string,
	profile *entity.UserProfile,
	prefs *entity.UserPreference,
	languageIDs *[]int32,
	interestIDs *[]int32,
	bumpStep bool,
	latitude *float64,
	longitude *float64,
) (err error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// ---- user_profiles (PATCH) ----
	if profile != nil {
		profile.UserID = userID

		upCols, insCols := profileColumnsProvided(profile)
		if len(insCols) == 0 && len(upCols) == 0 {
			// nothing provided -> skip
		} else {
			// Upsert with explicit column whitelists
			if err = profile.Upsert(
				ctx, tx,
				true,                       // update on conflict
				[]string{"user_id"},        // conflict target
				boil.Whitelist(upCols...),  // update columns (only provided)
				boil.Whitelist(insCols...), // insert columns (only provided + user_id)
			); err != nil {
				return fmt.Errorf("failed to upsert user_profile: %w", err)
			}
		}
	}

	if latitude != nil && longitude != nil {
		if _, err := tx.ExecContext(ctx, `
        UPDATE user_profiles
        SET geo = ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
            updated_at = now()
        WHERE user_id = $3
    `, *longitude, *latitude, userID); err != nil {
			return fmt.Errorf("failed to update geo: %w", err)
		}
	}

	// ---- user_preferences (PATCH) ----
	if prefs != nil {
		prefs.UserID = userID

		upCols, insCols := preferenceColumnsProvided(prefs)
		if len(insCols) == 0 && len(upCols) == 0 {
			// nothing provided -> skip
		} else {
			if err = prefs.Upsert(
				ctx, tx,
				true,
				[]string{"user_id"},
				boil.Whitelist(upCols...),
				boil.Whitelist(insCols...),
			); err != nil {
				return fmt.Errorf("failed to upsert user_preferences: %w", err)
			}
		}
	}

	// ---- Many-to-many replacements (only if provided) ----
	u, err := entity.FindUser(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if languageIDs != nil {
		langs, err := entity.Languages(
			qm.WhereIn("id IN ?", int16AnySlice(*languageIDs)...),
		).All(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to load languages: %w", err)
		}

		if err := u.SetLanguages(ctx, tx, false, langs...); err != nil {
			return fmt.Errorf("failed to set languages: %w", err)
		}
	}

	if interestIDs != nil {
		ints, err := entity.Interests(
			qm.WhereIn("id IN ?", int16AnySlice(*interestIDs)...),
		).All(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to load interests: %w", err)
		}

		if err := u.SetInterests(ctx, tx, false, ints...); err != nil {
			return fmt.Errorf("failed to set interests: %w", err)
		}
	}

	// ---- Optional: bump onboarding_step ----
	if bumpStep {
		u.OnboardingStep++
		if _, err = u.Update(ctx, tx, boil.Whitelist(entity.UserColumns.OnboardingStep)); err != nil {
			return fmt.Errorf("failed to update onboarding_step: %w", err)
		}
	}

	return nil
}

// Build update/insert column lists for PATCH semantics on user_profiles.
func profileColumnsProvided(p *entity.UserProfile) (updateCols, insertCols []string) {
	add := func(col string) { updateCols = append(updateCols, col); insertCols = append(insertCols, col) }

	// NOTE: only add columns that are actually provided (Valid=true or non-nil slice)
	if p.DisplayName.Valid {
		add(entity.UserProfileColumns.DisplayName)
	}

	if p.Bio.Valid {
		add(entity.UserProfileColumns.Bio)
	}

	if p.City.Valid {
		add(entity.UserProfileColumns.City)
	}

	if p.Country.Valid {
		add(entity.UserProfileColumns.Country)
	}

	if p.Work.Valid {
		add(entity.UserProfileColumns.Work)
	}

	if p.JobTitle.Valid {
		add(entity.UserProfileColumns.JobTitle)
	}

	if p.University.Valid {
		add(entity.UserProfileColumns.University)
	}

	if p.Birthdate.Valid {
		add(entity.UserProfileColumns.Birthdate)
	}

	if p.HeightCM.Valid {
		add(entity.UserProfileColumns.HeightCM)
	}

	if p.GenderID.Valid {
		add(entity.UserProfileColumns.GenderID)
	}

	if p.DatingIntentionID.Valid {
		add(entity.UserProfileColumns.DatingIntentionID)
	}

	if p.ReligionID.Valid {
		add(entity.UserProfileColumns.ReligionID)
	}

	if p.EducationLevelID.Valid {
		add(entity.UserProfileColumns.EducationLevelID)
	}

	if p.PoliticalBeliefID.Valid {
		add(entity.UserProfileColumns.PoliticalBeliefID)
	}

	if p.DrinkingID.Valid {
		add(entity.UserProfileColumns.DrinkingID)
	}

	if p.SmokingID.Valid {
		add(entity.UserProfileColumns.SmokingID)
	}

	if p.MarijuanaID.Valid {
		add(entity.UserProfileColumns.MarijuanaID)
	}

	if p.DrugsID.Valid {
		add(entity.UserProfileColumns.DrugsID)
	}

	if p.ChildrenStatusID.Valid {
		add(entity.UserProfileColumns.ChildrenStatusID)
	}

	if p.FamilyPlanID.Valid {
		add(entity.UserProfileColumns.FamilyPlanID)
	}

	if p.EthnicityID.Valid {
		add(entity.UserProfileColumns.EthnicityID)
	}

	// JSONB ([]byte) — only if provided (non-nil)
	if p.ProfileMeta.Valid {
		add(entity.UserProfileColumns.ProfileMeta)
	}

	// Always set updated_at if we’re touching the row
	if len(updateCols) > 0 {
		updateCols = append(updateCols, entity.UserProfileColumns.UpdatedAt)
	}
	// Ensure user_id included for inserts
	if len(insertCols) > 0 {
		insertCols = append(insertCols, entity.UserProfileColumns.UserID)
	}

	return
}

// Build update/insert column lists for PATCH semantics on user_preferences.
func preferenceColumnsProvided(pr *entity.UserPreference) (updateCols, insertCols []string) {
	add := func(col string) { updateCols = append(updateCols, col); insertCols = append(insertCols, col) }

	if pr.DistanceKM.Valid {
		add(entity.UserPreferenceColumns.DistanceKM)
	}

	if pr.AgeMin.Valid {
		add(entity.UserPreferenceColumns.AgeMin)
	}

	if pr.AgeMax.Valid {
		add(entity.UserPreferenceColumns.AgeMax)
	}

	// INT[] columns: include when slice is non-nil.
	if pr.SeekGenderIds != nil {
		add(entity.UserPreferenceColumns.SeekGenderIds)
	}

	if pr.SeekIntentionIds != nil {
		add(entity.UserPreferenceColumns.SeekIntentionIds)
	}

	if pr.SeekReligionIds != nil {
		add(entity.UserPreferenceColumns.SeekReligionIds)
	}

	if pr.SeekPoliticalBeliefIds != nil {
		add(entity.UserPreferenceColumns.SeekPoliticalBeliefIds)
	}

	if len(updateCols) > 0 {
		updateCols = append(updateCols, entity.UserPreferenceColumns.UpdatedAt)
	}

	if len(insertCols) > 0 {
		insertCols = append(insertCols, entity.UserPreferenceColumns.UserID)
	}

	return
}

func int16AnySlice(src []int32) []any {
	if len(src) == 0 {
		return []any{-1} // guard: IN () invalid
	}

	out := make([]any, len(src))
	for i, v := range src {
		out[i] = int16(v)
	}

	return out
}
