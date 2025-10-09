package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

type Service interface {
	GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error)
	GetProfileCard(ctx context.Context, userID string) (profilecard.ProfileCard, error)
	GetProfileForUpdate(ctx context.Context, userID string) (domain.UpdateProfile, error)
	UpdateProfile(ctx context.Context, updatedProfile domain.UpdateProfile) error
	ScaffoldProfile(ctx context.Context, tx *sql.Tx, userID string) error
	GetVoicePromptByID(ctx context.Context, id int64) (domain.VoicePrompt, error)

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

var (
	ErrContainsSocialMediaPromotion = fmt.Errorf("this field cannot contain social media promotion")
	ErrInvalidID                    = errors.New("id must be greater than 0")
	ErrMissingPrompts               = errors.New("missing prompts")
	ErrTooManyPromptsProvided       = errors.New("too many prompts provided")
	ErrInvalidBirthdate             = errors.New("invalid birthdate")
	ErrInvalidHeight                = errors.New("invalid height")
)

func (s *service) GetVoicePromptByID(ctx context.Context, id int64) (domain.VoicePrompt, error) {
	vp, err := s.profileRepo.GetVoicePromptByID(ctx, id)
	if err != nil {
		return domain.VoicePrompt{}, fmt.Errorf("get voice prompt: %w", err)
	}

	prompt, err := s.lookupRepo.GetPromptTypeByID(ctx, vp.PromptType.Int16)
	if err != nil {
		return domain.VoicePrompt{}, fmt.Errorf("get prompt type: %w", err)
	}

	var coverPhotoUrl string
	if vp.CoverPhotoURL.Valid {
		coverPhotoUrl = vp.CoverPhotoURL.String
	}

	return domain.VoicePrompt{
		PromptID:      vp.ID,
		VoiceNoteURL:  vp.AudioURL,
		CoverPhotoUrl: coverPhotoUrl,
		Prompt:        prompt.Label,
	}, nil
}

func (s *service) ScaffoldProfile(ctx context.Context, tx *sql.Tx, userID string) error {
	err := s.profileRepo.InsertProfile(ctx,
		&entity.UserProfile{
			UserID: userID,
			Emoji:  null.StringFrom(constants.DefaultEmoji),
		},
		tx,
	)
	if err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	return nil
}

func (s *service) GetProfileForUpdate(ctx context.Context, userID string) (domain.UpdateProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	languageIds, err := s.profileRepo.GetUserSpokenLanguages(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("get user spoken languages: %w", err)
	}

	VoicePrompts, err := s.getUserVoicePromptsForUpdate(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("get user voice prompts: %w", err)
	}

	Photos, err := s.getUserPhotos(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("get user photos: %w", err)
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
	if up.Birthdate != nil {
		bd := *up.Birthdate
		today := time.Now().UTC()

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
	if up.HeightCM != nil {
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

func (s *service) UpdateProfile(ctx context.Context, up domain.UpdateProfile) error {
	err := s.validateProfileUpdate(up)
	if err != nil {
		return fmt.Errorf("validate profile update: %w", err)
	}
	// Load current profile
	prof, err := s.getUserProfile(ctx, up.UserID)
	if err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	// Basic
	if up.DisplayName != nil {
		prof.DisplayName = *up.DisplayName
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

	// Persist
	err = s.updateUserProfile(ctx, prof)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	if up.BaseColour != nil {
		err = s.UpsertUserTheme(ctx, up.UserID, *up.BaseColour)
		if err != nil {
			return fmt.Errorf("upsert user theme: %w", err)
		}
	}

	if len(up.SpokenLanguages) > 0 {
		err = s.profileRepo.UpsertUserSpokenLanguages(ctx, up.UserID, up.SpokenLanguages)
		if err != nil {
			return fmt.Errorf("upsert user spoken languages: %w", err)
		}
	}

	if len(up.Photos) > 0 {
		err = s.profileRepo.UpsertUserPhotos(ctx, up.UserID, mapper.MapUpdatedPhotosToEntity(up.Photos, up.UserID))
		if err != nil {
			return fmt.Errorf("insert user photos: %w", err)
		}
	}

	if len(up.VoicePrompts) > 0 {
		err = s.profileRepo.UpsertUserPrompts(ctx, up.UserID, mapper.MapVoicePromptsUpdateToEntity(up.VoicePrompts, up.UserID))
		if err != nil {
			return fmt.Errorf("insert user voice prompts: %w", err)
		}
	}

	return nil
}

func (s *service) GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	result := domain.EnrichedProfile{
		DisplayName:   userProfile.DisplayName,
		Birthdate:     userProfile.Birthdate,
		Age:           utils.CalculateAge(userProfile.Birthdate),
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
		return domain.EnrichedProfile{}, fmt.Errorf("get user theme: %w", err)
	}

	result.Gender, err = s.getGenderByID(ctx, userProfile.GenderID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get gender: %w", err)
	}

	result.Ethnicity, err = s.getEthnicityByID(ctx, userProfile.EthnicityID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get ethnicity: %w", err)
	}

	result.DatingIntention, err = s.getDatingIntentionByID(ctx, userProfile.DatingIntentionID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get dating intention: %w", err)
	}

	result.Religion, err = s.getReligionByID(ctx, userProfile.ReligionID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get religion: %w", err)
	}

	result.EducationLevel, err = s.getEducationLevelByID(ctx, userProfile.EducationLevelID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get education level: %w", err)
	}

	result.PoliticalBelief, err = s.getPoliticalBeliefByID(ctx, userProfile.PoliticalBeliefID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get political belief: %w", err)
	}

	result.Drinking, err = s.getHabitByID(ctx, userProfile.DrinkingID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get drinking habit: %w", err)
	}

	result.Smoking, err = s.getHabitByID(ctx, userProfile.SmokingID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get smoking habit: %w", err)
	}

	result.Marijuana, err = s.getHabitByID(ctx, userProfile.MarijuanaID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get marijuana habit: %w", err)
	}

	result.Drugs, err = s.getHabitByID(ctx, userProfile.DrugsID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get drugs habit: %w", err)
	}

	if userProfile.ChildrenStatusID != nil {
		result.ChildrenStatus, err = s.getFamilyStatusByID(ctx, *userProfile.ChildrenStatusID)
		if err != nil {
			return domain.EnrichedProfile{}, fmt.Errorf("get children status: %w", err)
		}
	}

	if userProfile.FamilyPlanID != nil {
		result.FamilyPlan, err = s.getFamilyPlanByID(ctx, *userProfile.FamilyPlanID)
		if err != nil {
			return domain.EnrichedProfile{}, fmt.Errorf("get family plan: %w", err)
		}
	}

	result.SpokenLanguages, err = s.getUserSpokenLanguages(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user spoken languages: %w", err)
	}

	result.VoicePrompts, err = s.getUserVoicePrompts(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user voice prompts: %w", err)
	}

	result.Photos, err = s.getUserPhotos(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user photos: %w", err)
	}

	return result, nil
}

func (s *service) GetProfileCard(ctx context.Context, userID string) (profilecard.ProfileCard, error) {
	enrichedProfile, err := s.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return profilecard.ProfileCard{}, fmt.Errorf("get enriched profile: %w", err)
	}

	return mapper.MapEnrichedProfileToProfileCard(enrichedProfile), nil
}

func (s *service) UpsertUserTheme(ctx context.Context, userID, baseColour string) error {
	// generate colours
	palJSON, err := s.generatePaletteJsonFromBaseColour(baseColour)
	if err != nil {
		return fmt.Errorf("generate palette json: %w", err)
	}
	// store colours.
	err = s.profileRepo.UpsertUserTheme(ctx, entity.UserTheme{
		UserID:  userID,
		BaseHex: baseColour,
		Palette: palJSON,
	})
	if err != nil {
		return fmt.Errorf("upsert user theme: %w", err)
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

	return nil
}

func (s *service) UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error {
	if err := validateUserPromptsUpsert(prompts); err != nil {
		return fmt.Errorf("validate user prompts: %w", err)
	}
	// todo: check if position values are unique
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
