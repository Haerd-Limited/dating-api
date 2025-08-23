package storage

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type OnboardingRepository interface {
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
	_, err := userProfile.Update(ctx, or.db, boil.Whitelist(
		entity.UserProfileColumns.DisplayName,
		entity.UserProfileColumns.Birthdate,
		entity.UserProfileColumns.HeightCM,
		entity.UserProfileColumns.City,
		entity.UserProfileColumns.Country,
		entity.UserProfileColumns.GenderID,
		entity.UserProfileColumns.DatingIntentionID,
		entity.UserProfileColumns.ReligionID,
		entity.UserProfileColumns.EducationLevelID,
		entity.UserProfileColumns.PoliticalBeliefID,
		entity.UserProfileColumns.DrinkingID,
		entity.UserProfileColumns.SmokingID,
		entity.UserProfileColumns.MarijuanaID,
		entity.UserProfileColumns.DrugsID,
		entity.UserProfileColumns.ChildrenStatusID,
		entity.UserProfileColumns.FamilyPlanID,
		entity.UserProfileColumns.EthnicityID,
		entity.UserProfileColumns.Work,
		entity.UserProfileColumns.JobTitle,
		entity.UserProfileColumns.University,
		entity.UserProfileColumns.ProfileMeta,
		entity.UserProfileColumns.UpdatedAt,
	))
	if err != nil {
		return err
	}

	return nil
}
