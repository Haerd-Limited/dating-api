package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/openai"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	verificationstorage "github.com/Haerd-Limited/dating-api/internal/verification/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

type Service interface {
	GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error)
	GetProfileCard(ctx context.Context, userID string) (profilecard.ProfileCard, error)
	GetProfileCardWithDistance(ctx context.Context, userID string, currentUserLat, currentUserLon float64) (profilecard.ProfileCard, error)
	GetProfileForUpdate(ctx context.Context, userID string) (domain.UpdateProfile, error)
	UpdateProfile(ctx context.Context, updatedProfile domain.UpdateProfile) error
	ScaffoldProfile(ctx context.Context, tx *sql.Tx, userID string) error
	GetVoicePromptByID(ctx context.Context, id int64) (domain.VoicePrompt, error)
	GetUserPhotos(ctx context.Context, userID string) ([]domain.Photo, error)
	GetTranscript(ctx context.Context, voicePromptID int64) (string, error)

	UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error
	UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error
	UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error
	VerifyProfile(ctx context.Context, userID string) error
	IsVerified(ctx context.Context, userID string) (bool, error)
	SetProfileUnderReview(ctx context.Context, userID string) error
	SetProfileUnverified(ctx context.Context, userID string) error
	// Stats
	CountBasicsCompletedByGender(ctx context.Context, genderID int16) (int64, error)
	CountBasicsCompleted(ctx context.Context) (int64, error)
}

type service struct {
	logger           *zap.Logger
	profileRepo      storage.ProfileRepository
	lookupRepo       lookupstorage.LookupRepository
	verificationRepo verificationstorage.VerificationRepository
	openaiService    openai.Service
	awsService       aws.Service
	uow              uow.UoW
}

func NewProfileService(
	logger *zap.Logger,
	profileRepository storage.ProfileRepository,
	lookupRepository lookupstorage.LookupRepository,
	verificationRepository verificationstorage.VerificationRepository,
	openaiService openai.Service,
	awsService aws.Service,
	uow uow.UoW,
) Service {
	return &service{
		logger:           logger,
		profileRepo:      profileRepository,
		lookupRepo:       lookupRepository,
		verificationRepo: verificationRepository,
		openaiService:    openaiService,
		awsService:       awsService,
		uow:              uow,
	}
}

var (
	ErrContainsSocialMediaPromotion = fmt.Errorf("this field cannot contain social media promotion")
	ErrInvalidID                    = errors.New("id must be greater than 0")
	ErrMissingPrompts               = errors.New("missing prompts")
	ErrTooManyPromptsProvided       = errors.New("too many prompts provided")
	ErrNotEnoughPhotosProvided      = errors.New("not enough photos provided")
	ErrInvalidBirthdate             = errors.New("invalid birthdate")
	ErrInvalidHeight                = errors.New("invalid height")
	ErrInvalidPromptPosition        = errors.New("invalid prompt position")
	ErrDuplicatePromptPosition      = errors.New("duplicate prompt position")
	ErrDuplicatePhotoPosition       = errors.New("duplicate photo position")
)

func (s *service) VerifyProfile(ctx context.Context, userID string) error {
	// Load current profile
	prof, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", userID))
	}

	prof.VerifiedStatus = domain.VerificationStatusVerified

	err = s.updateUserProfile(ctx, prof, nil)
	if err != nil {
		return commonlogger.LogError(s.logger, "update user profile", err, zap.String("userID", userID))
	}

	return nil
}

func (s *service) IsVerified(ctx context.Context, userID string) (bool, error) {
	status, err := s.profileRepo.GetVerificationStatus(ctx, userID)
	if err != nil {
		return false, err
	}

	return status == domain.VerificationStatusVerified, nil
}

func (s *service) CountBasicsCompletedByGender(ctx context.Context, genderID int16) (int64, error) {
	count, err := s.profileRepo.CountUsersBasicsCompletedByGender(ctx, genderID)
	if err != nil {
		return 0, commonlogger.LogError(s.logger, "count basics-completed by gender", err, zap.Int16("genderID", genderID))
	}

	return count, nil
}

func (s *service) CountBasicsCompleted(ctx context.Context) (int64, error) {
	count, err := s.profileRepo.CountUsersBasicsCompleted(ctx)
	if err != nil {
		return 0, commonlogger.LogError(s.logger, "count basics-completed", err)
	}

	return count, nil
}

func (s *service) GetVoicePromptByID(ctx context.Context, id int64) (domain.VoicePrompt, error) {
	vp, err := s.profileRepo.GetVoicePromptByID(ctx, id)
	if err != nil {
		return domain.VoicePrompt{}, fmt.Errorf("get voice prompt: %w", err)
	}

	waveformData, err := mapper.UnmarshalWaveformData(vp.WaveformData)
	if err != nil {
		return domain.VoicePrompt{}, fmt.Errorf("unmarshal waveform data: %w", err)
	}

	prompt, err := s.lookupRepo.GetPromptTypeByID(ctx, vp.PromptType.Int16)
	if err != nil {
		return domain.VoicePrompt{}, fmt.Errorf("get prompt type: %w", err)
	}

	var coverMediaURL string
	if vp.CoverMediaURL.Valid {
		coverMediaURL = vp.CoverMediaURL.String
	}

	var coverMediaType *string
	if vp.CoverMediaType.Valid {
		coverMediaType = &vp.CoverMediaType.String
	}

	var coverMediaAspectRatio *float64

	if vp.CoverMediaAspectRatio.Valid {
		aspectRatio64 := float64(vp.CoverMediaAspectRatio.Float32)
		coverMediaAspectRatio = &aspectRatio64
	}

	return domain.VoicePrompt{
		PromptID:              vp.ID,
		VoiceNoteURL:          vp.AudioURL,
		WaveformData:          waveformData,
		CoverMediaURL:         coverMediaURL,
		CoverMediaType:        coverMediaType,
		CoverMediaAspectRatio: coverMediaAspectRatio,
		Prompt:                prompt.Label,
	}, nil
}

func (s *service) GetTranscript(ctx context.Context, voicePromptID int64) (string, error) {
	// 1. Get voice prompt from DB
	vp, err := s.profileRepo.GetVoicePromptByID(ctx, voicePromptID)
	if err != nil {
		return "", fmt.Errorf("voice prompt not found: %w", err)
	}

	// 2. Return if transcript already exists
	if vp.Transcript.Valid && vp.Transcript.String != "" {
		return vp.Transcript.String, nil
	}

	// 3. Extract S3 key from audio_url
	key, err := utils.S3KeyFromURL(vp.AudioURL)
	if err != nil {
		return "", fmt.Errorf("invalid audio URL: %w", err)
	}

	// 4. Download audio from S3
	audioData, err := s.awsService.GetObjectBytes(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to download audio: %w", err)
	}

	// 5. Transcribe with OpenAI
	transcript, err := s.openaiService.TranscribeAudio(ctx, audioData, key)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	// 6. Save to database
	err = s.profileRepo.UpdateVoicePromptTranscript(ctx, voicePromptID, transcript)
	if err != nil {
		s.logger.Warn("failed to save transcript", zap.Error(err))
		// Don't fail - return the transcript anyway
	}

	return transcript, nil
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
		return domain.UpdateProfile{}, commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", userID))
	}

	languageIds, err := s.profileRepo.GetUserSpokenLanguages(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, commonlogger.LogError(s.logger, "get user spoken languages", err, zap.String("userID", userID))
	}

	ethnicityIDs, err := s.profileRepo.GetUserEthnicities(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, commonlogger.LogError(s.logger, "get user ethnicities", err, zap.String("userID", userID))
	}

	VoicePrompts, err := s.getUserVoicePromptsForUpdate(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, commonlogger.LogError(s.logger, "get user voice prompts", err, zap.String("userID", userID))
	}

	Photos, err := s.getUserPhotos(ctx, userID)
	if err != nil {
		return domain.UpdateProfile{}, commonlogger.LogError(s.logger, "get user photos", err, zap.String("userID", userID))
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
		Work:              userProfile.Work,
		JobTitle:          userProfile.JobTitle,
		University:        userProfile.University,
		CreatedAt:         &userProfile.CreatedAt,
		UpdatedAt:         userProfile.UpdatedAt,
	}, nil
}

func (s *service) UpdateProfile(ctx context.Context, up domain.UpdateProfile) error {
	err := s.validateProfileUpdate(up)
	if err != nil {
		return commonlogger.LogError(s.logger, "validate profile update", err, zap.String("userID", up.UserID))
	}

	// Begin transaction
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin tx", err, zap.String("userID", up.UserID))
	}

	defer func() { _ = tx.Rollback() }()

	// Load current profile
	prof, err := s.getUserProfile(ctx, up.UserID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", up.UserID))
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

	if up.SexualityID != nil {
		prof.SexualityID = *up.SexualityID
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
		err = s.profileRepo.UpsertUserPhotos(ctx, up.UserID, mapper.MapUpdatedPhotosToEntity(up.Photos, up.UserID), tx.Raw())
		if err != nil {
			return commonlogger.LogError(s.logger, "insert user photos", err, zap.String("userID", up.UserID))
		}

		err = s.verificationRepo.InvalidatePhotoVerification(ctx, up.UserID, tx.Raw())
		if err != nil {
			return commonlogger.LogError(s.logger, "invalidate photo verification", err, zap.String("userID", up.UserID))
		}

		prof.VerifiedStatus = domain.VerificationStatusUnverified
	}

	// Persist
	err = s.updateUserProfile(ctx, prof, tx.Raw())
	if err != nil {
		return commonlogger.LogError(s.logger, "update user profile", err, zap.String("userID", up.UserID))
	}

	// Update ethnicities if provided (including empty array to clear)
	if up.EthnicityIDs != nil {
		err = s.profileRepo.UpsertUserEthnicities(ctx, up.UserID, up.EthnicityIDs, tx.Raw())
		if err != nil {
			return commonlogger.LogError(s.logger, "upsert user ethnicities", err, zap.String("userID", up.UserID))
		}
	}

	if len(up.SpokenLanguages) > 0 {
		err = s.profileRepo.UpsertUserSpokenLanguages(ctx, up.UserID, up.SpokenLanguages, tx.Raw())
		if err != nil {
			return commonlogger.LogError(s.logger, "upsert user spoken languages", err, zap.String("userID", up.UserID))
		}
	}

	if len(up.VoicePrompts) > 0 {
		marshalledWaveformData, err := marshalVoicePromptsForEntity(up.VoicePrompts)
		if err != nil {
			return commonlogger.LogError(s.logger, "marshal user voice prompts", err, zap.String("userID", up.UserID))
		}

		err = s.profileRepo.UpsertUserPrompts(ctx, up.UserID, mapper.MapVoicePromptsUpdateToEntity(up.VoicePrompts, marshalledWaveformData, up.UserID), tx.Raw())
		if err != nil {
			return commonlogger.LogError(s.logger, "insert user voice prompts", err, zap.String("userID", up.UserID))
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return commonlogger.LogError(s.logger, "commit tx", err, zap.String("userID", up.UserID))
	}

	return nil
}

func (s *service) GetEnrichedProfile(ctx context.Context, userID string) (domain.EnrichedProfile, error) {
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	result := domain.EnrichedProfile{
		DisplayName:    userProfile.DisplayName,
		Birthdate:      userProfile.Birthdate,
		Age:            utils.CalculateAge(userProfile.Birthdate),
		HeightCM:       userProfile.HeightCM,
		UserID:         userID,
		Latitude:       userProfile.Latitude,
		Longitude:      userProfile.Longitude,
		City:           userProfile.City,
		Country:        userProfile.Country,
		Work:           userProfile.Work,
		JobTitle:       userProfile.JobTitle,
		University:     userProfile.University,
		CreatedAt:      userProfile.CreatedAt,
		UpdatedAt:      userProfile.UpdatedAt,
		Emoji:          userProfile.Emoji,
		VerifiedStatus: userProfile.VerifiedStatus,
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

	result.Sexuality, err = s.getSexualityByID(ctx, userProfile.SexualityID)
	if err != nil {
		return domain.EnrichedProfile{}, fmt.Errorf("get sexuality: %w", err)
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

func (s *service) GetProfileCardWithDistance(ctx context.Context, userID string, currentUserLat, currentUserLon float64) (profilecard.ProfileCard, error) {
	enrichedProfile, err := s.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return profilecard.ProfileCard{}, fmt.Errorf("get enriched profile: %w", err)
	}

	profileCard := mapper.MapEnrichedProfileToProfileCard(enrichedProfile)

	// Calculate distance between current user and profile
	profileCard.DistanceKm = utils.CalculateDistanceKm(currentUserLat, currentUserLon, enrichedProfile.Latitude, enrichedProfile.Longitude)

	return profileCard, nil
}

func (s *service) UpsertUserPrompts(ctx context.Context, userID string, prompts []domain.VoicePromptUpdate) error {
	if err := validateUserPromptsUpsert(prompts); err != nil {
		return fmt.Errorf("validate user prompts: %w", err)
	}

	marshalledWaveformData, err := marshalVoicePromptsForEntity(prompts)
	if err != nil {
		return fmt.Errorf("marshal voice prompts: %w", err)
	}

	return s.profileRepo.UpsertUserPrompts(ctx, userID, mapper.MapVoicePromptsUpdateToEntity(prompts, marshalledWaveformData, userID), nil)
}

func (s *service) UpsertUserPhotos(ctx context.Context, userID string, photos []domain.Photo) error {
	if len(photos) != 6 {
		return fmt.Errorf("%w: please provide exactly 6 photos", ErrNotEnoughPhotosProvided)
	}

	// Check if position values are unique
	seen := make(map[int16]struct{}, len(photos))
	for _, p := range photos {
		if _, dup := seen[p.Position]; dup {
			return fmt.Errorf("%w: duplicate position=%d", ErrDuplicatePhotoPosition, p.Position)
		}

		seen[p.Position] = struct{}{}
	}

	// Begin transaction
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin tx", err, zap.String("userID", userID))
	}

	defer func() { _ = tx.Rollback() }()

	// Upsert photos
	err = s.profileRepo.UpsertUserPhotos(ctx, userID, mapper.MapUpdatedPhotosToEntity(photos, userID), tx.Raw())
	if err != nil {
		return commonlogger.LogError(s.logger, "upsert user photos", err, zap.String("userID", userID))
	}

	// Invalidate photo verification
	err = s.verificationRepo.InvalidatePhotoVerification(ctx, userID, tx.Raw())
	if err != nil {
		return commonlogger.LogError(s.logger, "invalidate photo verification", err, zap.String("userID", userID))
	}

	// Load current profile and set verification status to unverified
	prof, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", userID))
	}

	prof.VerifiedStatus = domain.VerificationStatusUnverified

	err = s.updateUserProfile(ctx, prof, tx.Raw())
	if err != nil {
		return commonlogger.LogError(s.logger, "update user profile", err, zap.String("userID", userID))
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return commonlogger.LogError(s.logger, "commit tx", err, zap.String("userID", userID))
	}

	return nil
}

func (s *service) UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error {
	return s.profileRepo.UpsertUserSpokenLanguages(ctx, userID, languages, nil)
}

func (s *service) SetProfileUnderReview(ctx context.Context, userID string) error {
	// Load current profile
	prof, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", userID))
	}

	prof.VerifiedStatus = domain.VerificationStatusUnderReview

	err = s.updateUserProfile(ctx, prof, nil)
	if err != nil {
		return commonlogger.LogError(s.logger, "update user profile", err, zap.String("userID", userID))
	}

	return nil
}

func (s *service) SetProfileUnverified(ctx context.Context, userID string) error {
	// Load current profile
	prof, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get user profile", err, zap.String("userID", userID))
	}

	prof.VerifiedStatus = domain.VerificationStatusUnverified

	err = s.updateUserProfile(ctx, prof, nil)
	if err != nil {
		return commonlogger.LogError(s.logger, "update user profile", err, zap.String("userID", userID))
	}

	return nil
}
