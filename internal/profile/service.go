package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
	verificationstorage "github.com/Haerd-Limited/dating-api/internal/verification/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
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
	GetUserPhotos(ctx context.Context, userID string) ([]domain.Photo, error)

	UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error
	UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error
	UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error
	UpsertUserTheme(ctx context.Context, userID, baseColour string) error
	VerifyProfile(ctx context.Context, userID string) error
}

type service struct {
	logger           *zap.Logger
	profileRepo      storage.ProfileRepository
	lookupRepo       lookupstorage.LookupRepository
	verificationRepo verificationstorage.VerificationRepository
}

func NewProfileService(
	logger *zap.Logger,
	profileRepository storage.ProfileRepository,
	lookupRepository lookupstorage.LookupRepository,
	verificationRepository verificationstorage.VerificationRepository,
) Service {
	return &service{
		logger:           logger,
		profileRepo:      profileRepository,
		lookupRepo:       lookupRepository,
		verificationRepo: verificationRepository,
	}
}

var (
	ErrContainsSocialMediaPromotion = fmt.Errorf("this field cannot contain social media promotion")
	ErrInvalidID                    = errors.New("id must be greater than 0")
	ErrMissingPrompts               = errors.New("missing prompts")
	ErrTooManyPromptsProvided       = errors.New("too many prompts provided")
	ErrInvalidBirthdate             = errors.New("invalid birthdate")
	ErrInvalidHeight                = errors.New("invalid height")
	ErrInvalidPromptPosition        = errors.New("invalid prompt position")
	ErrDuplicatePromptPosition      = errors.New("duplicate prompt position")
)

func (s *service) VerifyProfile(ctx context.Context, userID string) error {
	// Load current profile
	prof, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	prof.Verified = true

	err = s.updateUserProfile(ctx, prof)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return nil
}

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

	ethnicityIDs, err := s.profileRepo.GetUserEthnicities(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, fmt.Errorf("get user ethnicities: %w", err)
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
		EthnicityIDs:      ethnicityIDs,
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

// todo(high-priority): make atomic with uow
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

	if len(up.Photos) > 0 {
		err = s.profileRepo.UpsertUserPhotos(ctx, up.UserID, mapper.MapUpdatedPhotosToEntity(up.Photos, up.UserID))
		if err != nil {
			return fmt.Errorf("insert user photos: %w", err)
		}

		err = s.verificationRepo.InvalidatePhotoVerification(ctx, up.UserID)
		if err != nil {
			return fmt.Errorf("invalidate photo verification: %w", err)
		}

		prof.Verified = false
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

	// Update ethnicities if provided (including empty array to clear)
	if up.EthnicityIDs != nil {
		err = s.profileRepo.UpsertUserEthnicities(ctx, up.UserID, up.EthnicityIDs)
		if err != nil {
			return fmt.Errorf("upsert user ethnicities: %w", err)
		}
	}

	if len(up.SpokenLanguages) > 0 {
		err = s.profileRepo.UpsertUserSpokenLanguages(ctx, up.UserID, up.SpokenLanguages)
		if err != nil {
			return fmt.Errorf("upsert user spoken languages: %w", err)
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
		Verified:      userProfile.Verified,
	}

	result.Theme, err = s.getUserTheme(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user theme: %w", err)
	}

	result.Gender, err = s.getGenderByID(ctx, userProfile.GenderID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get gender: %w", err)
	}

	ethnicityIDs, err := s.profileRepo.GetUserEthnicities(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user ethnicities: %w", err)
	}

	result.Ethnicities, err = s.getEthnicitiesByIDs(ctx, ethnicityIDs)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get ethnicities: %w", err)
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

func (s *service) GetUserPhotos(ctx context.Context, userID string) ([]domain.Photo, error) {
	return s.getUserPhotos(ctx, userID)
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

func (s *service) UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error {
	if err := validateUserPromptsUpsert(prompts); err != nil {
		return fmt.Errorf("validate user prompts: %w", err)
	}

	return s.profileRepo.UpsertUserPrompts(ctx, userID, mapper.MapVoicePromptsUpdateToEntity(prompts, userID))
}

func (s *service) UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error {
	// todo(high-priority): check if position values are unique
	// todo(high-priority): ensure count is min/max 6
	return s.profileRepo.UpsertUserPhotos(ctx, userID, mapper.MapUpdatedPhotosToEntity(photos, userID))
}

func (s *service) UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error {
	return s.profileRepo.UpsertUserSpokenLanguages(ctx, userID, languages)
}
