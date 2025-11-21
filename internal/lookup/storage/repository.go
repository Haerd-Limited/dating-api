package storage

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type LookupRepository interface {
	GetFamilyStatusByID(ctx context.Context, id int16) (*entity.FamilyStatus, error)
	GetFamilyPlanByID(ctx context.Context, id int16) (*entity.FamilyPlan, error)
	GetLanguageByID(ctx context.Context, id int16) (*entity.Language, error)
	GetEducationLevelByID(ctx context.Context, id int16) (*entity.EducationLevel, error)
	GetEthnicityByID(ctx context.Context, id int16) (*entity.Ethnicity, error)
	GetReligionByID(ctx context.Context, id int16) (*entity.Religion, error)
	GetPoliticalBeliefByID(ctx context.Context, id int16) (*entity.PoliticalBelief, error)
	GetHabitByID(ctx context.Context, id int16) (*entity.Habit, error)
	GetDatingIntentionByID(ctx context.Context, id int16) (*entity.DatingIntention, error)
	GetGenderByID(ctx context.Context, id int16) (*entity.Gender, error)
	GetSexualityByID(ctx context.Context, id int16) (*entity.Sexuality, error)
	GetPromptTypeByID(ctx context.Context, id int16) (*entity.PromptType, error)
	GetLanguages(ctx context.Context) (entity.LanguageSlice, error)
	GetEducationLevels(ctx context.Context) (entity.EducationLevelSlice, error)
	GetEthnicities(ctx context.Context) (entity.EthnicitySlice, error)
	GetReligions(ctx context.Context) (entity.ReligionSlice, error)
	GetPoliticalBeliefs(ctx context.Context) (entity.PoliticalBeliefSlice, error)
	GetHabits(ctx context.Context) (entity.HabitSlice, error)
	GetDatingIntentions(ctx context.Context) (entity.DatingIntentionSlice, error)
	GetGenders(ctx context.Context) (entity.GenderSlice, error)
	GetPrompts(ctx context.Context) (entity.PromptTypeSlice, error)
	GetFamilyPlans(ctx context.Context) (entity.FamilyPlanSlice, error)
	GetFamilyStatus(ctx context.Context) (entity.FamilyStatusSlice, error)
	GetReportCategories(ctx context.Context) (entity.ReportCategorySlice, error)
}

type lookupRepository struct {
	db *sqlx.DB
}

func NewLookupRepository(db *sqlx.DB) LookupRepository {
	return &lookupRepository{
		db: db,
	}
}

func (lr *lookupRepository) GetFamilyStatus(ctx context.Context) (entity.FamilyStatusSlice, error) {
	familyStatus, err := entity.FamilyStatuses().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return familyStatus, nil
}

func (lr *lookupRepository) GetReportCategories(ctx context.Context) (entity.ReportCategorySlice, error) {
	reportCategories, err := entity.ReportCategories(
		qm.OrderBy("sort_order ASC, id ASC"),
	).All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return reportCategories, nil
}

func (lr *lookupRepository) GetFamilyPlans(ctx context.Context) (entity.FamilyPlanSlice, error) {
	familyPlans, err := entity.FamilyPlans().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return familyPlans, nil
}

func (lr *lookupRepository) GetPrompts(ctx context.Context) (entity.PromptTypeSlice, error) {
	prompts, err := entity.PromptTypes().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return prompts, nil
}

func (lr *lookupRepository) GetLanguages(ctx context.Context) (entity.LanguageSlice, error) {
	languages, err := entity.Languages().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return languages, nil
}

func (lr *lookupRepository) GetEducationLevels(ctx context.Context) (entity.EducationLevelSlice, error) {
	educationLevels, err := entity.EducationLevels().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return educationLevels, nil
}

func (lr *lookupRepository) GetEthnicities(ctx context.Context) (entity.EthnicitySlice, error) {
	ethnicities, err := entity.Ethnicities().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return ethnicities, nil
}

func (lr *lookupRepository) GetReligions(ctx context.Context) (entity.ReligionSlice, error) {
	religions, err := entity.Religions().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return religions, nil
}

func (lr *lookupRepository) GetPoliticalBeliefs(ctx context.Context) (entity.PoliticalBeliefSlice, error) {
	politicalBeliefs, err := entity.PoliticalBeliefs().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return politicalBeliefs, nil
}

func (lr *lookupRepository) GetHabits(ctx context.Context) (entity.HabitSlice, error) {
	habits, err := entity.Habits().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return habits, nil
}

func (lr *lookupRepository) GetDatingIntentions(ctx context.Context) (entity.DatingIntentionSlice, error) {
	di, err := entity.DatingIntentions().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return di, nil
}

func (lr *lookupRepository) GetGenders(ctx context.Context) (entity.GenderSlice, error) {
	genders, err := entity.Genders().All(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return genders, nil
}

func (lr *lookupRepository) GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error) {
	userProfile, err := entity.UserProfiles(entity.UserProfileWhere.UserID.EQ(userID)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (lr *lookupRepository) GetLanguageByID(ctx context.Context, id int16) (*entity.Language, error) {
	language, err := entity.Languages(entity.LanguageWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return language, nil
}

func (lr *lookupRepository) GetEducationLevelByID(ctx context.Context, id int16) (*entity.EducationLevel, error) {
	educationLevel, err := entity.EducationLevels(entity.EducationLevelWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return educationLevel, nil
}

func (lr *lookupRepository) GetEthnicityByID(ctx context.Context, id int16) (*entity.Ethnicity, error) {
	ethnicity, err := entity.Ethnicities(entity.EthnicityWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return ethnicity, nil
}

func (lr *lookupRepository) GetReligionByID(ctx context.Context, id int16) (*entity.Religion, error) {
	religion, err := entity.Religions(entity.ReligionWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return religion, nil
}

func (lr *lookupRepository) GetPoliticalBeliefByID(ctx context.Context, id int16) (*entity.PoliticalBelief, error) {
	politicalBelief, err := entity.PoliticalBeliefs(entity.PoliticalBeliefWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return politicalBelief, nil
}

func (lr *lookupRepository) GetHabitByID(ctx context.Context, id int16) (*entity.Habit, error) {
	habit, err := entity.Habits(entity.HabitWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (lr *lookupRepository) GetDatingIntentionByID(ctx context.Context, id int16) (*entity.DatingIntention, error) {
	datingIntention, err := entity.DatingIntentions(entity.DatingIntentionWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return datingIntention, nil
}

func (lr *lookupRepository) GetGenderByID(ctx context.Context, id int16) (*entity.Gender, error) {
	gender, err := entity.Genders(entity.GenderWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return gender, nil
}

func (lr *lookupRepository) GetSexualityByID(ctx context.Context, id int16) (*entity.Sexuality, error) {
	sexuality, err := entity.Sexualities(entity.SexualityWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return sexuality, nil
}

func (lr *lookupRepository) GetPromptTypeByID(ctx context.Context, id int16) (*entity.PromptType, error) {
	promptType, err := entity.PromptTypes(entity.PromptTypeWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return promptType, nil
}

func (lr *lookupRepository) GetFamilyStatusByID(ctx context.Context, id int16) (*entity.FamilyStatus, error) {
	familyStatus, err := entity.FamilyStatuses(entity.FamilyStatusWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return familyStatus, nil
}

func (lr *lookupRepository) GetFamilyPlanByID(ctx context.Context, id int16) (*entity.FamilyPlan, error) {
	familyPlan, err := entity.FamilyPlans(entity.FamilyPlanWhere.ID.EQ(id)).One(ctx, lr.db)
	if err != nil {
		return nil, err
	}

	return familyPlan, nil
}
