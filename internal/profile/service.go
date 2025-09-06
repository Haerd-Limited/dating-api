package profile

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
)

type Service interface {
	GetMyProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error)
}

type service struct {
	logger      *zap.Logger
	profileRepo storage.ProfileRepository
	lookupRepo  lookupstorage.LookupRepository
}

func NewProfileService(
	logger *zap.Logger,
	profileRepository storage.ProfileRepository,
	lookupRepository lookupstorage.LookupRepository,
) Service {
	return &service{
		logger:      logger,
		profileRepo: profileRepository,
		lookupRepo:  lookupRepository,
	}
}

func (s *service) GetMyProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error) {
	userProfile, err := s.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, err
	}

	return userProfile, nil
}

func (s *service) GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user profile: %w", err)
	}

	result := domain.EnrichedProfile{
		DisplayName: userProfile.DisplayName,
		Birthdate:   userProfile.Birthdate,
		Age:         calculateAge(userProfile.Birthdate),
		HeightCM:    userProfile.HeightCM,
		UserID:      userID,
		Latitude:    userProfile.Latitude,
		Longitude:   userProfile.Longitude,
		City:        userProfile.City,
		Country:     userProfile.Country,
		Work:        userProfile.Work,
		JobTitle:    userProfile.JobTitle,
		University:  userProfile.University,
		CreatedAt:   userProfile.CreatedAt,
		UpdatedAt:   userProfile.UpdatedAt,
	}

	result.Gender, err = s.getGenderByID(ctx, userProfile.GenderID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get gender: %w", err)
	}

	result.Ethnicity, err = s.getEthnicityByID(ctx, userProfile.EthnicityID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get ethnicity: %w", err)
	}

	result.DatingIntention, err = s.getDatingIntentionByID(ctx, userProfile.DatingIntentionID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get dating intention: %w", err)
	}

	result.Religion, err = s.getReligionByID(ctx, userProfile.ReligionID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get religion: %w", err)
	}

	result.EducationLevel, err = s.getEducationLevelByID(ctx, userProfile.EducationLevelID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get education level: %w", err)
	}

	result.PoliticalBelief, err = s.getPoliticalBeliefByID(ctx, userProfile.PoliticalBeliefID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get political belief: %w", err)
	}

	result.Drinking, err = s.getHabitByID(ctx, userProfile.DrinkingID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get drinking habit: %w", err)
	}

	result.Smoking, err = s.getHabitByID(ctx, userProfile.SmokingID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get smoking habit: %w", err)
	}

	result.Marijuana, err = s.getHabitByID(ctx, userProfile.MarijuanaID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get marijuana habit: %w", err)
	}

	result.Drugs, err = s.getHabitByID(ctx, userProfile.DrugsID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get drugs habit: %w", err)
	}

	if userProfile.ChildrenStatusID != nil {
		result.ChildrenStatus, err = s.getFamilyStatusByID(ctx, *userProfile.ChildrenStatusID)
		if err != nil {
			return domain.EnrichedProfile{}, fmt.Errorf("failed to get children status: %w", err)
		}
	}

	if userProfile.FamilyPlanID != nil {
		result.FamilyPlan, err = s.getFamilyPlanByID(ctx, *userProfile.FamilyPlanID)
		if err != nil {
			return domain.EnrichedProfile{}, fmt.Errorf("failed to get family plan: %w", err)
		}
	}

	result.SpokenLanguages, err = s.getUserSpokenLanguages(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user spoken languages: %w", err)
	}

	result.VoicePrompts, err = s.getUserVoicePrompts(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user voice prompts: %w", err)
	}

	result.Photos, err = s.getUserPhotos(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user photos: %w", err)
	}

	return result, nil
}

func (s *service) getUserPhotos(ctx context.Context, userID string) ([]domain.Photo, error) {
	photos, err := s.profileRepo.GetUserPhotos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user photos: %w", err)
	}

	var photosList []domain.Photo

	for _, photo := range photos {
		photosList = append(photosList, domain.Photo{
			URL:       photo.URL,
			IsPrimary: photo.IsPrimary,
			Position:  photo.Position.Int16,
		})
	}

	return photosList, nil
}

func (s *service) getUserVoicePrompts(ctx context.Context, userID string) ([]domain.VoicePrompt, error) {
	voicePromptEntities, err := s.profileRepo.GetUserVoicePrompts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user voice prompts: %w", err)
	}

	var voicePrompts []domain.VoicePrompt

	for _, vpe := range voicePromptEntities {
		if !vpe.PromptType.Valid {
			return nil, fmt.Errorf("invalid prompt type: promptType=%v", vpe.PromptType.Int16)
		}

		var vpeErr error

		promptType, vpeErr := s.lookupRepo.GetPromptTypeByID(ctx, vpe.PromptType.Int16)
		if vpeErr != nil {
			return nil, fmt.Errorf("failed to get prompt type by ID: %w", vpeErr)
		}

		voicePrompts = append(voicePrompts, domain.VoicePrompt{
			URL: vpe.AudioURL,
			PromptType: domain.Prompt{
				ID:       promptType.ID,
				Label:    promptType.Label,
				Key:      promptType.Key,
				Category: promptType.Category,
			},
			IsPrimary:  vpe.IsPrimary,
			Position:   vpe.Position.Int16,
			DurationMs: vpe.DurationMS,
		})
	}

	return voicePrompts, nil
}

func (s *service) getUserProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	userProfileEntity, err := s.profileRepo.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapper.MapProfileToDomain(userProfileEntity), nil
}

func (s *service) getUserSpokenLanguages(ctx context.Context, userID string) ([]domain.Language, error) {
	languageIds, err := s.profileRepo.GetUserSpokenLanguages(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user spoken languages: %w", err)
	}

	var languages []domain.Language

	for _, languageID := range languageIds {
		var langErr error

		languageEntity, langErr := s.lookupRepo.GetLanguageByID(ctx, languageID)
		if langErr != nil {
			return nil, fmt.Errorf("failed to get language by ID: %w", langErr)
		}

		languages = append(languages, domain.Language{
			ID:    languageEntity.ID,
			Label: languageEntity.Label,
		})
	}

	return languages, nil
}

func (s *service) getGenderByID(ctx context.Context, genderID int16) (domain.Gender, error) {
	genderEntity, err := s.lookupRepo.GetGenderByID(ctx, genderID)
	if err != nil {
		return domain.Gender{}, err
	}

	return domain.Gender{
		ID:    genderEntity.ID,
		Label: genderEntity.Label,
	}, nil
}

func (s *service) getDatingIntentionByID(ctx context.Context, datingIntentionID int16) (domain.DatingIntention, error) {
	datingIntentionEntity, err := s.lookupRepo.GetDatingIntentionByID(ctx, datingIntentionID)
	if err != nil {
		return domain.DatingIntention{}, err
	}

	return domain.DatingIntention{
		ID:    datingIntentionEntity.ID,
		Label: datingIntentionEntity.Label,
	}, nil
}

func (s *service) getReligionByID(ctx context.Context, religionID int16) (domain.Religion, error) {
	religionEntity, err := s.lookupRepo.GetReligionByID(ctx, religionID)
	if err != nil {
		return domain.Religion{}, err
	}

	return domain.Religion{
		ID:    religionEntity.ID,
		Label: religionEntity.Label,
	}, nil
}

func (s *service) getEducationLevelByID(ctx context.Context, educationLevelID int16) (domain.EducationLevel, error) {
	educationLevelEntity, err := s.lookupRepo.GetEducationLevelByID(ctx, educationLevelID)
	if err != nil {
		return domain.EducationLevel{}, err
	}

	return domain.EducationLevel{
		ID:    educationLevelEntity.ID,
		Label: educationLevelEntity.Label,
	}, nil
}

func (s *service) getPoliticalBeliefByID(ctx context.Context, politicalBeliefID int16) (domain.PoliticalBelief, error) {
	politicalBeliefEntity, err := s.lookupRepo.GetPoliticalBeliefByID(ctx, politicalBeliefID)
	if err != nil {
		return domain.PoliticalBelief{}, err
	}

	return domain.PoliticalBelief{
		ID:    politicalBeliefEntity.ID,
		Label: politicalBeliefEntity.Label,
	}, err
}

func (s *service) getHabitByID(ctx context.Context, habitID int16) (domain.Habit, error) {
	habitEntity, err := s.lookupRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return domain.Habit{}, err
	}

	return domain.Habit{
		ID:    habitEntity.ID,
		Label: habitEntity.Label,
	}, nil
}

func (s *service) getFamilyStatusByID(ctx context.Context, id int16) (*domain.Status, error) {
	statusEntity, err := s.lookupRepo.GetFamilyStatusByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if statusEntity == nil {
		return nil, nil
	}

	return &domain.Status{
		ID:    statusEntity.ID,
		Label: statusEntity.Label,
		Key:   statusEntity.Key.String,
	}, nil
}

func (s *service) getFamilyPlanByID(ctx context.Context, id int16) (*domain.Status, error) {
	statusEntity, err := s.lookupRepo.GetFamilyPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if statusEntity == nil {
		return nil, nil
	}

	return &domain.Status{
		ID:    statusEntity.ID,
		Label: statusEntity.Label,
		Key:   statusEntity.Key.String,
	}, nil
}

func (s *service) getEthnicityByID(ctx context.Context, id int16) (domain.Ethnicity, error) {
	ethnicityEntity, err := s.lookupRepo.GetEthnicityByID(ctx, id)
	if err != nil {
		return domain.Ethnicity{}, err
	}

	if ethnicityEntity == nil {
		return domain.Ethnicity{}, nil
	}

	return domain.Ethnicity{
		ID:    ethnicityEntity.ID,
		Label: ethnicityEntity.Label,
	}, nil
}

// calculateAge returns the age in years given a birthdate.
func calculateAge(birthdate time.Time) int {
	now := time.Now()

	years := now.Year() - birthdate.Year()

	// If the birthday hasn't occurred yet this year, subtract 1
	if now.Month() < birthdate.Month() ||
		(now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		years--
	}

	return years
}
