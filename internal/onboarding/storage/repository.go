package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface {
	InsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error
	GetLanguages(ctx context.Context) (entity.LanguageSlice, error)
	GetEducationLevels(ctx context.Context) (entity.EducationLevelSlice, error)
	GetEthnicities(ctx context.Context) (entity.EthnicitySlice, error)
	GetReligions(ctx context.Context) (entity.ReligionSlice, error)
	GetPoliticalBeliefs(ctx context.Context) (entity.PoliticalBeliefSlice, error)
	GetHabits(ctx context.Context) (entity.HabitSlice, error)
	GetDatingIntentions(ctx context.Context) (entity.DatingIntentionSlice, error)
	GetGenders(ctx context.Context) (entity.GenderSlice, error)
	GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error)
	UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile) error
}

type onboardingRepository struct {
	db *sqlx.DB
}

func NewOnboardingRepository(db *sqlx.DB) OnboardingRepository {
	return &onboardingRepository{
		db: db,
	}
}

func (or *onboardingRepository) InsertUserSpokenLanguages(
	ctx context.Context,
	userID string,
	languages []int16,
) (err error) {
	tx, err := or.db.BeginTxx(ctx, &sql.TxOptions{})
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

	// Clear existing selections
	_, err = tx.ExecContext(ctx, `DELETE FROM user_languages WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user_languages: %w", err)
	}

	// Nothing to insert? we're done after clearing.
	if len(languages) == 0 {
		return nil
	}

	// De‑dupe in case the caller passed duplicates.
	uniq := make([]int16, 0, len(languages))

	seen := make(map[int16]struct{}, len(languages))
	for _, id := range languages {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}

		uniq = append(uniq, id)
	}

	// Build a single INSERT ... VALUES (...) ... statement.
	var (
		sb   strings.Builder
		args = make([]any, 1+len(uniq))
	)

	sb.WriteString(`INSERT INTO user_languages (user_id, language_id) VALUES `)

	args[0] = userID

	for i, lid := range uniq {
		if i > 0 {
			sb.WriteString(",")
		}
		// user_id is always $1; each language_id is $2, $3, ...
		sb.WriteString(fmt.Sprintf("($1,$%d)", i+2))

		args[i+1] = lid
	}

	// Ignore duplicates that could arise from race conditions.
	sb.WriteString(` ON CONFLICT (user_id, language_id) DO NOTHING`)

	if _, err = tx.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert user_languages: %w", err)
	}

	return nil
}

func (or *onboardingRepository) GetLanguages(ctx context.Context) (entity.LanguageSlice, error) {
	languages, err := entity.Languages().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return languages, nil
}

func (or *onboardingRepository) GetEducationLevels(ctx context.Context) (entity.EducationLevelSlice, error) {
	educationLevels, err := entity.EducationLevels().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return educationLevels, nil
}

func (or *onboardingRepository) GetEthnicities(ctx context.Context) (entity.EthnicitySlice, error) {
	ethnicities, err := entity.Ethnicities().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return ethnicities, nil
}

func (or *onboardingRepository) GetReligions(ctx context.Context) (entity.ReligionSlice, error) {
	religions, err := entity.Religions().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return religions, nil
}

func (or *onboardingRepository) GetPoliticalBeliefs(ctx context.Context) (entity.PoliticalBeliefSlice, error) {
	politicalBeliefs, err := entity.PoliticalBeliefs().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return politicalBeliefs, nil
}

func (or *onboardingRepository) GetHabits(ctx context.Context) (entity.HabitSlice, error) {
	habits, err := entity.Habits().All(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return habits, nil
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

func (or *onboardingRepository) GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error) {
	userProfile, err := entity.UserProfiles(entity.UserProfileWhere.UserID.EQ(userID)).One(ctx, or.db)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (or *onboardingRepository) UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile) error {
	_, err := userProfile.Update(ctx, or.db, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}
