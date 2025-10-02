package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/theme"
)

type Service interface {
	GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error)
	GetProfileCard(ctx context.Context, userID string) (profilecard.ProfileCard, error)
	GetProfileForUpdate(ctx context.Context, userID string) (domain.UpdateProfile, error)
	UpdateProfile(ctx context.Context, updatedProfile domain.UpdateProfile) error

	UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error
	UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error
	UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error
	UpsertUserTheme(ctx context.Context, userID, baseColour string) error
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

func (s *service) GetProfileForUpdate(ctx context.Context, userID string) (domain.UpdateProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("failed to get user profile: %w", err)
	}

	languageIds, err := s.profileRepo.GetUserSpokenLanguages(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("failed to get user spoken languages: %w", err)
	}

	VoicePrompts, err := s.getUserVoicePromptsForUpdate(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("failed to get user voice prompts: %w", err)
	}

	Photos, err := s.getUserPhotos(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("failed to get user photos: %w", err)
	}

	return domain.UpdateProfile{
		DisplayName:       &userProfile.DisplayName,
		Birthdate:         &userProfile.Birthdate,
		HeightCM:          &userProfile.HeightCM,
		UserID:            userProfile.UserID,
		ProfileEmoji:      &userProfile.Emoji,
		Latitude:          &userProfile.Latitude,
		Longitude:         &userProfile.Longitude,
		City:              &userProfile.City,
		Country:           &userProfile.Country,
		GenderID:          &userProfile.GenderID,
		DatingIntentionID: &userProfile.DatingIntentionID,
		ReligionID:        &userProfile.ReligionID,
		EducationLevelID:  &userProfile.EducationLevelID,
		PoliticalBeliefID: &userProfile.PoliticalBeliefID,
		DrinkingID:        &userProfile.DrinkingID,
		SmokingID:         &userProfile.SmokingID,
		MarijuanaID:       &userProfile.MarijuanaID,
		DrugsID:           &userProfile.DrugsID,
		ChildrenStatusID:  userProfile.ChildrenStatusID,
		FamilyPlanID:      userProfile.FamilyPlanID,
		EthnicityID:       &userProfile.EthnicityID,
		SpokenLanguages:   languageIds,
		VoicePrompts:      VoicePrompts,
		Photos:            Photos,
		CoverPhotoURL:     userProfile.CoverPhotoURL,
		Work:              userProfile.Work,
		JobTitle:          userProfile.JobTitle,
		University:        userProfile.University,
		CreatedAt:         &userProfile.CreatedAt,
		UpdatedAt:         userProfile.UpdatedAt,
	}, nil
}

func (s *service) UpdateProfile(ctx context.Context, up domain.UpdateProfile) error {
	// Load current profile
	prof, err := s.getUserProfile(ctx, up.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	// Basic
	if up.DisplayName != nil {
		prof.DisplayName = *up.DisplayName
	}

	if up.Birthdate != nil {
		prof.Birthdate = *up.Birthdate
	}

	if up.HeightCM != nil {
		prof.HeightCM = *up.HeightCM
	}

	if up.ProfileEmoji != nil {
		prof.Emoji = *up.ProfileEmoji
	}

	if up.CoverPhotoURL != nil {
		prof.CoverPhotoURL = up.CoverPhotoURL
	}

	// Location
	if up.Latitude != nil {
		prof.Latitude = *up.Latitude
	}

	if up.Longitude != nil {
		prof.Longitude = *up.Longitude
	}

	if up.City != nil {
		prof.City = *up.City
	}

	if up.Country != nil {
		prof.Country = *up.Country
	}

	if up.GenderID != nil {
		prof.GenderID = *up.GenderID
	}

	if up.DatingIntentionID != nil {
		prof.DatingIntentionID = *up.DatingIntentionID
	}

	if up.ReligionID != nil {
		prof.ReligionID = *up.ReligionID
	}

	if up.EducationLevelID != nil {
		prof.EducationLevelID = *up.EducationLevelID
	}

	if up.PoliticalBeliefID != nil {
		prof.PoliticalBeliefID = *up.PoliticalBeliefID
	}

	if up.DrinkingID != nil {
		prof.DrinkingID = *up.DrinkingID
	}

	if up.SmokingID != nil {
		prof.SmokingID = *up.SmokingID
	}

	if up.MarijuanaID != nil {
		prof.MarijuanaID = *up.MarijuanaID
	}

	if up.DrugsID != nil {
		prof.DrugsID = *up.DrugsID
	}

	if up.ChildrenStatusID != nil {
		prof.ChildrenStatusID = up.ChildrenStatusID
	}

	if up.FamilyPlanID != nil {
		prof.FamilyPlanID = up.FamilyPlanID
	}

	if up.EthnicityID != nil {
		prof.EthnicityID = *up.EthnicityID
	}

	// Work / education text fields
	if up.Work != nil {
		prof.Work = up.Work
	}

	if up.JobTitle != nil {
		prof.JobTitle = up.JobTitle
	}

	if up.University != nil {
		prof.University = up.University
	}

	// Persist
	err = s.updateUserProfile(ctx, prof)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	if up.BaseColour != nil {
		err = s.UpsertUserTheme(ctx, up.UserID, *up.BaseColour)
		if err != nil {
			return fmt.Errorf("failed to upsert user theme: %w", err)
		}
	}

	if len(up.SpokenLanguages) > 0 {
		err = s.profileRepo.UpsertUserSpokenLanguages(ctx, up.UserID, up.SpokenLanguages)
		if err != nil {
			return fmt.Errorf("failed to upsert user spoken languages: %w", err)
		}
	}

	if len(up.Photos) > 0 {
		err = s.profileRepo.UpsertUserPhotos(ctx, up.UserID, mapper.MapUpdatedPhotosToEntity(up.Photos, up.UserID))
		if err != nil {
			return fmt.Errorf("failed to insert user photos: %w", err)
		}
	}

	if len(up.VoicePrompts) > 0 {
		err = s.profileRepo.UpsertUserPrompts(ctx, up.UserID, mapper.MapVoicePromptsUpdateToEntity(up.VoicePrompts, up.UserID))
		if err != nil {
			return fmt.Errorf("failed to insert user voice prompts: %w", err)
		}
	}

	return nil
}

func (s *service) updateUserProfile(ctx context.Context, userProfile *domain.Profile) error {
	updatedUserProfileEntity, whitelist, err := mapper.MapProfileToEntityForUpdate(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = s.profileRepo.UpdateUserProfile(ctx, updatedUserProfileEntity, whitelist)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (s *service) GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user profile: %w", err)
	}

	result := domain.EnrichedProfile{
		DisplayName:   userProfile.DisplayName,
		Birthdate:     userProfile.Birthdate,
		Age:           calculateAge(userProfile.Birthdate),
		HeightCM:      userProfile.HeightCM,
		UserID:        userID,
		Latitude:      userProfile.Latitude,
		Longitude:     userProfile.Longitude,
		City:          userProfile.City,
		Country:       userProfile.Country,
		Work:          userProfile.Work,
		JobTitle:      userProfile.JobTitle,
		University:    userProfile.University,
		CreatedAt:     userProfile.CreatedAt,
		UpdatedAt:     userProfile.UpdatedAt,
		CoverPhotoURL: userProfile.CoverPhotoURL,
		Emoji:         userProfile.Emoji,
	}

	result.Theme, err = s.getUserTheme(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("failed to get user theme: %w", err)
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

func (s *service) GetProfileCard(ctx context.Context, userID string) (profilecard.ProfileCard, error) {
	enrichedProfile, err := s.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return profilecard.ProfileCard{}, fmt.Errorf("failed to get enriched profile: %w", err)
	}

	return mapper.MapEnrichedProfileToProfileCard(enrichedProfile), nil
}

func (s *service) UpsertUserTheme(ctx context.Context, userID, baseColour string) error {
	// generate colours
	palJSON, err := s.generatePaletteJsonFromBaseColour(baseColour)
	if err != nil {
		return fmt.Errorf("failed to generate palette json: %w", err)
	}
	// store colours.
	err = s.profileRepo.UpsertUserTheme(ctx, entity.UserTheme{
		UserID:  userID,
		BaseHex: baseColour,
		Palette: palJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert user theme: %w", err)
	}

	return nil
}

func (s *service) generatePaletteJsonFromBaseColour(baseColour string) ([]byte, error) {
	palette, err := theme.GeneratePalette9(baseColour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	palJSON, err := json.Marshal(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}

	return palJSON, nil
}

func (s *service) UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error {
	// todo: check if position values are unique
	// todo: ensure count is max 6
	return s.profileRepo.UpsertUserPrompts(ctx, userID, mapper.MapVoicePromptsUpdateToEntity(prompts, userID))
}

func (s *service) UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error {
	// todo: check if position values are unique
	// todo: ensure count is min/max 6
	return s.profileRepo.UpsertUserPhotos(ctx, userID, mapper.MapUpdatedPhotosToEntity(photos, userID))
}

func (s *service) UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error {
	return s.profileRepo.UpsertUserSpokenLanguages(ctx, userID, languages)
}

func (s *service) getUserTheme(ctx context.Context, userID string) (domain.UserTheme, error) {
	userThemeEntity, err := s.profileRepo.GetUserTheme(ctx, userID)
	if err != nil {
		return domain.UserTheme{}, fmt.Errorf("failed to get user theme: %w", err)
	}

	if userThemeEntity == nil {
		return domain.UserTheme{}, nil
	}

	result := domain.UserTheme{
		BaseHex: userThemeEntity.BaseHex,
	}

	err = userThemeEntity.Palette.Unmarshal(&result.Palette)
	if err != nil {
		return domain.UserTheme{}, fmt.Errorf("failed to unmarshal palette: %w", err)
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

		var promptCoverURL string
		if vpe.CoverPhotoURL.Valid {
			promptCoverURL = vpe.CoverPhotoURL.String
		}

		voicePrompts = append(voicePrompts, domain.VoicePrompt{
			URL: vpe.AudioURL,
			PromptType: domain.Prompt{
				ID:       promptType.ID,
				Label:    promptType.Label,
				Key:      promptType.Key,
				Category: promptType.Category,
			},
			IsPrimary:      vpe.IsPrimary,
			Position:       vpe.Position.Int16,
			DurationMs:     vpe.DurationMS,
			PromptCoverURL: promptCoverURL,
		})
	}

	return voicePrompts, nil
}

func (s *service) getUserVoicePromptsForUpdate(ctx context.Context, userID string) ([]domain.VoicePromptUpdate, error) {
	vps, err := s.getUserVoicePrompts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user voice prompts: %w", err)
	}

	var result []domain.VoicePromptUpdate
	for _, vp := range vps {
		result = append(result, domain.VoicePromptUpdate{
			URL:            vp.URL,
			PromptTypeID:   vp.PromptType.ID,
			IsPrimary:      vp.IsPrimary,
			Position:       vp.Position,
			DurationMs:     vp.DurationMs,
			PromptCoverURL: vp.PromptCoverURL,
		})
	}

	return result, nil
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
		return domain.Gender{}, fmt.Errorf("failed to get gender by ID: %w", err)
	}

	if genderEntity == nil {
		return domain.Gender{}, errors.New("gender not found")
	}

	return domain.Gender{
		ID:    genderEntity.ID,
		Label: genderEntity.Label,
	}, nil
}

func (s *service) getDatingIntentionByID(ctx context.Context, datingIntentionID int16) (domain.DatingIntention, error) {
	datingIntentionEntity, err := s.lookupRepo.GetDatingIntentionByID(ctx, datingIntentionID)
	if err != nil {
		return domain.DatingIntention{}, fmt.Errorf("failed to get dating intention by ID: %w", err)
	}

	if datingIntentionEntity == nil {
		return domain.DatingIntention{}, errors.New("dating intention not found")
	}

	return domain.DatingIntention{
		ID:    datingIntentionEntity.ID,
		Label: datingIntentionEntity.Label,
	}, nil
}

func (s *service) getReligionByID(ctx context.Context, religionID int16) (domain.Religion, error) {
	religionEntity, err := s.lookupRepo.GetReligionByID(ctx, religionID)
	if err != nil {
		return domain.Religion{}, fmt.Errorf("failed to get religion by ID: %w", err)
	}

	if religionEntity == nil {
		return domain.Religion{}, errors.New("religion not found")
	}

	return domain.Religion{
		ID:    religionEntity.ID,
		Label: religionEntity.Label,
	}, nil
}

func (s *service) getEducationLevelByID(ctx context.Context, educationLevelID int16) (domain.EducationLevel, error) {
	educationLevelEntity, err := s.lookupRepo.GetEducationLevelByID(ctx, educationLevelID)
	if err != nil {
		return domain.EducationLevel{}, fmt.Errorf("failed to get education level by ID: %w", err)
	}

	if educationLevelEntity == nil {
		return domain.EducationLevel{}, errors.New("education level not found")
	}

	return domain.EducationLevel{
		ID:    educationLevelEntity.ID,
		Label: educationLevelEntity.Label,
	}, nil
}

func (s *service) getPoliticalBeliefByID(ctx context.Context, politicalBeliefID int16) (domain.PoliticalBelief, error) {
	politicalBeliefEntity, err := s.lookupRepo.GetPoliticalBeliefByID(ctx, politicalBeliefID)
	if err != nil {
		return domain.PoliticalBelief{}, fmt.Errorf("failed to get political belief by ID: %w", err)
	}

	if politicalBeliefEntity == nil {
		return domain.PoliticalBelief{}, errors.New("political belief not found")
	}

	return domain.PoliticalBelief{
		ID:    politicalBeliefEntity.ID,
		Label: politicalBeliefEntity.Label,
	}, err
}

func (s *service) getHabitByID(ctx context.Context, habitID int16) (domain.Habit, error) {
	habitEntity, err := s.lookupRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return domain.Habit{}, fmt.Errorf("failed to get habit by ID: %w", err)
	}

	if habitEntity == nil {
		return domain.Habit{}, errors.New("habit not found")
	}

	return domain.Habit{
		ID:    habitEntity.ID,
		Label: habitEntity.Label,
	}, nil
}

func (s *service) getFamilyStatusByID(ctx context.Context, id int16) (*domain.Status, error) {
	statusEntity, err := s.lookupRepo.GetFamilyStatusByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get family status by ID: %w", err)
	}

	if statusEntity == nil {
		return nil, errors.New("family status not found")
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
		return nil, fmt.Errorf("failed to get family plan by ID: %w", err)
	}

	if statusEntity == nil {
		return nil, errors.New("family plan not found")
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
		return domain.Ethnicity{}, fmt.Errorf("failed to get ethnicity by ID: %w", err)
	}

	if ethnicityEntity == nil {
		return domain.Ethnicity{}, errors.New("ethnicity not found")
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
