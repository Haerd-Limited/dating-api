package lookup

import (
	"context"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/lookup/domain"
	"github.com/Haerd-Limited/dating-api/internal/lookup/mapper"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
)

type Service interface {
	GetPrompts(ctx context.Context) ([]domain.Prompt, error)
	GetFamilyPlans(ctx context.Context) ([]domain.FamilyPlan, error)
	GetLanguages(ctx context.Context) ([]domain.Language, error)
	GetReligions(ctx context.Context) ([]domain.Religion, error)
	GetPoliticalBeliefs(ctx context.Context) ([]domain.PoliticalBelief, error)
	GetEducationLevels(ctx context.Context) ([]domain.EducationLevel, error)
	GetEthnicities(ctx context.Context) ([]domain.Ethnicity, error)
	GetHabits(ctx context.Context) ([]domain.Habit, error)
	GetGenders(ctx context.Context) ([]domain.Gender, error)
	GetDatingIntentions(ctx context.Context) ([]domain.DatingIntention, error)
}

type lookupService struct {
	logger     *zap.Logger
	lookupRepo lookupstorage.LookupRepository
}

func NewLookupService(
	logger *zap.Logger,
	lookupRepo lookupstorage.LookupRepository,
) Service {
	return &lookupService{
		logger:     logger,
		lookupRepo: lookupRepo,
	}
}

func (s *lookupService) GetFamilyPlans(ctx context.Context) ([]domain.FamilyPlan, error) {
	familyPlanEntities, err := s.lookupRepo.GetFamilyPlans(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapFamilyPlanEntitiesToDomain(familyPlanEntities), nil
}

func (s *lookupService) GetPrompts(ctx context.Context) ([]domain.Prompt, error) {
	prompts, err := s.lookupRepo.GetPrompts(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPromptsToDomain(prompts), nil
}

func (s *lookupService) GetLanguages(ctx context.Context) ([]domain.Language, error) {
	languageEntities, err := s.lookupRepo.GetLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapLanguagesToDomain(languageEntities), nil
}

func (s *lookupService) GetReligions(ctx context.Context) ([]domain.Religion, error) {
	religionsEntities, err := s.lookupRepo.GetReligions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapReligionsToDomain(religionsEntities), nil
}

func (s *lookupService) GetPoliticalBeliefs(ctx context.Context) ([]domain.PoliticalBelief, error) {
	politicalBeliefsEntities, err := s.lookupRepo.GetPoliticalBeliefs(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPoliticalBeliefsToDomain(politicalBeliefsEntities), nil
}

func (s *lookupService) GetEducationLevels(ctx context.Context) ([]domain.EducationLevel, error) {
	educationLevelEntities, err := s.lookupRepo.GetEducationLevels(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEducationLevelsToDomain(educationLevelEntities), nil
}

func (s *lookupService) GetEthnicities(ctx context.Context) ([]domain.Ethnicity, error) {
	ethnicityEntities, err := s.lookupRepo.GetEthnicities(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEthnicityToDomain(ethnicityEntities), nil
}

func (s *lookupService) GetHabits(ctx context.Context) ([]domain.Habit, error) {
	habitEntities, err := s.lookupRepo.GetHabits(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapHabitsToDomain(habitEntities), nil
}

func (s *lookupService) GetGenders(ctx context.Context) ([]domain.Gender, error) {
	genderEntities, err := s.lookupRepo.GetGenders(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapGendersToDomain(genderEntities), nil
}

func (s *lookupService) GetDatingIntentions(ctx context.Context) ([]domain.DatingIntention, error) {
	datingIntentionsEntities, err := s.lookupRepo.GetDatingIntentions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapDatingIntentionsToDomain(datingIntentionsEntities), nil
}
