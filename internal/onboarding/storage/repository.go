package storage

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface {
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
