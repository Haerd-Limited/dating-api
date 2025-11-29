package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/profile/constant"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/theme"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

func (s *service) validateProfileUpdate(up domain.UpdateProfile) error {
	if up.DisplayName != nil {
		if s.containsSocialMediaPromotion(*up.DisplayName) {
			return fmt.Errorf("%w : field=display_name, value=%s", ErrContainsSocialMediaPromotion, *up.DisplayName)
		}
	}

	if up.Work != nil {
		if s.containsSocialMediaPromotion(*up.Work) {
			return fmt.Errorf("%w : field=work, value=%s", ErrContainsSocialMediaPromotion, *up.Work)
		}
	}

	if up.JobTitle != nil {
		if s.containsSocialMediaPromotion(*up.JobTitle) {
			return fmt.Errorf("%w : field=work, value=%s", ErrContainsSocialMediaPromotion, *up.JobTitle)
		}
	}

	if up.University != nil {
		if s.containsSocialMediaPromotion(*up.University) {
			return fmt.Errorf("%w : field=university, value=%s", ErrContainsSocialMediaPromotion, *up.University)
		}
	}
	// birthdate
	if up.Birthdate != nil && !up.Birthdate.IsZero() {
		bd := *up.Birthdate
		today := time.Now()

		if bd.After(today) {
			return fmt.Errorf("%w: birthdate in future", ErrInvalidBirthdate)
		}

		age := utils.CalculateAge(bd)
		if age < constants.MinAge {
			return fmt.Errorf("%w: must be 18+", ErrInvalidBirthdate)
		}

		if age > constants.MaxAge {
			return fmt.Errorf("%w: must be realistic age", ErrInvalidBirthdate)
		}
	}

	// height
	if up.HeightCM != nil && *up.HeightCM != 0 {
		h := *up.HeightCM
		if h < constants.MinHeight || h > constants.MaxHeight {
			return fmt.Errorf("%w: height_cm out of range", ErrInvalidHeight)
		}
	}

	// URL
	if up.CoverPhotoURL != nil {
		if err := utils.ValidateHTTPURL(*up.CoverPhotoURL); err != nil {
			return fmt.Errorf("%w: cover_photo_url invalid: %v", commonErrors.ErrInvalidMediaUrl, err)
		}
		// Optional: enforce your CDN domain
		// if !strings.HasSuffix(u.Host, "your-cdn.com") { ... }
	}

	return nil
}

func validateUserPromptsUpsert(prompts []domain.VoicePromptUpdate) error {
	if len(prompts) == 0 {
		return ErrMissingPrompts
	}

	if len(prompts) > constants.MaximumNumberOfPrompts {
		return fmt.Errorf("%w. please provide atmost %v", ErrTooManyPromptsProvided, constants.MaximumNumberOfPrompts)
	}

	// positions: unique, 1..max
	seen := make(map[int16]struct{}, len(prompts))

	for i, p := range prompts {
		if p.Position <= 0 || p.Position > constants.MaximumNumberOfPrompts {
			return fmt.Errorf("%w: item[%d] position=%d must be 1..%d",
				ErrInvalidPromptPosition, i, p.Position, constants.MaximumNumberOfPrompts)
		}

		if _, dup := seen[p.Position]; dup {
			return fmt.Errorf("%w: duplicate position=%d", ErrDuplicatePromptPosition, p.Position)
		}

		seen[p.Position] = struct{}{}
	}

	return nil
}

func (s *service) containsSocialMediaPromotion(input string) bool {
	for _, token := range constant.BlockedTokens {
		if strings.Contains(strings.ToLower(input), token) {
			s.logger.Warn("containsSocialMediaPromotion", zap.String("input", input), zap.String("token", token))
			return true
		}
	}

	return false
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

func (s *service) updateUserProfile(ctx context.Context, userProfile *domain.Profile, tx *sql.Tx) error {
	updatedUserProfileEntity, whitelist, err := mapper.MapProfileToEntityForUpdate(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = s.profileRepo.UpdateUserProfile(ctx, updatedUserProfileEntity, whitelist, tx)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
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

func (s *service) getUserVoicePrompts(ctx context.Context, userID string) ([]domain.ProfileVoicePrompt, error) {
	voicePromptEntities, err := s.profileRepo.GetUserVoicePrompts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user voice prompts: %w", err)
	}

	var voicePrompts []domain.ProfileVoicePrompt

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

		voicePrompts = append(voicePrompts, domain.ProfileVoicePrompt{
			ID:  vpe.ID,
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

func (s *service) getSexualityByID(ctx context.Context, sexualityID int16) (domain.Sexuality, error) {
	sexualityEntity, err := s.lookupRepo.GetSexualityByID(ctx, sexualityID)
	if err != nil {
		return domain.Sexuality{}, fmt.Errorf("failed to get sexuality by ID: %w", err)
	}

	if sexualityEntity == nil {
		return domain.Sexuality{}, errors.New("sexuality not found")
	}

	return domain.Sexuality{
		ID:    sexualityEntity.ID,
		Label: sexualityEntity.Label,
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

func (s *service) getEthnicitiesByIDs(ctx context.Context, ids []int16) ([]domain.Ethnicity, error) {
	if len(ids) == 0 {
		return []domain.Ethnicity{}, nil
	}

	var ethnicities []domain.Ethnicity

	for _, id := range ids {
		ethnicity, err := s.getEthnicityByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get ethnicity by ID %d: %w", id, err)
		}

		ethnicities = append(ethnicities, ethnicity)
	}

	return ethnicities, nil
}
