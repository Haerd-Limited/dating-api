package onboarding

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	onboardingstorage "github.com/Haerd-Limited/dating-api/internal/onboarding/storage"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

type Service interface {
	Register(ctx context.Context, registerDetails domain.Register) (domain.RegisterResult, error)
	Basics(ctx context.Context, basicDetails domain.Basics) (domain.BasicsResult, error)
	Location(ctx context.Context, locationDetails domain.Location) (domain.LocationResult, error)
	Lifestyle(ctx context.Context, lifestyleDetails domain.Lifestyle) (domain.LifestyleResult, error)
	Beliefs(ctx context.Context, beliefDetails domain.Beliefs) (domain.BeliefsResult, error)
	Background(ctx context.Context, backgroundDetails domain.Background) (domain.BackgroundResult, error)
	WorkAndEducation(ctx context.Context, waeDetails domain.WorkAndEducation) (domain.WorkAndEducationResult, error)
	Languages(ctx context.Context, spokenLanguages domain.Languages) (domain.LanguagesResult, error)
	Photos(ctx context.Context, uploadedPhotos domain.UploadedPhotos) (domain.PhotosResult, error)
	Prompts(ctx context.Context, uploadedPrompts domain.Prompts) (domain.PromptsResult, error)
}

type onboardingService struct {
	logger      *zap.Logger
	repo        onboardingstorage.OnboardingRepository
	userService user.Service
	authService auth.Service
	awsService  aws.Service
}

func NewOnboardingService(
	logger *zap.Logger,
	onboardingRepository onboardingstorage.OnboardingRepository,
	userService user.Service,
	authService auth.Service,
	awsService aws.Service,
) Service {
	return &onboardingService{
		logger:      logger,
		repo:        onboardingRepository,
		userService: userService,
		authService: authService,
		awsService:  awsService,
	}
}

func (os *onboardingService) Register(ctx context.Context, registerDetails domain.Register) (domain.RegisterResult, error) {
	// Call user service and insert into user service
	userID, err := os.userService.CreateUser(ctx, userdomain.User{
		Email:          registerDetails.Email,
		PhoneNumber:    registerDetails.PhoneNumber,
		FirstName:      registerDetails.FirstName,
		LastName:       registerDetails.LastName,
		OnboardingStep: string(domain.OnboardingStepsBasics),
	})
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to create user: %w", err)
	}
	// generate and store access and refresh tokens
	tokens, err := os.authService.GenerateAccessAndRefreshToken(ctx, *userID)
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to generate access and refresh tokens: %w", err)
	}
	// Get dating intentions and genders and populate Content
	genders, err := os.getGenders(ctx)
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to get genders: %w", err)
	}

	datingIntentions, err := os.getDatingIntentions(ctx)
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to get dating intentions: %w", err)
	}

	return domain.RegisterResult{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		OnboardingSteps: domain.OnboardingStepsBasics.GenerateOnboardingSteps(),
		Content: domain.RegisterContent{
			DatingIntentions: datingIntentions,
			Genders:          genders,
		},
	}, nil
}

// TODO: FIX hitting the same endpoint multiple times advancing steps
func (os *onboardingService) Basics(ctx context.Context, basicDetails domain.Basics) (domain.BasicsResult, error) {
	userProfile, err := os.getUserProfile(ctx, basicDetails.UserID)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.GenderID = basicDetails.GenderID
	userProfile.DatingIntentionID = basicDetails.DatingIntentionID
	userProfile.HeightCM = basicDetails.HeightCm
	userProfile.Birthdate = basicDetails.Birthdate

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, basicDetails.UserID)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.BasicsResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) Location(ctx context.Context, locationDetails domain.Location) (domain.LocationResult, error) {
	userProfile, err := os.getUserProfile(ctx, locationDetails.UserID)
	if err != nil {
		return domain.LocationResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.City = locationDetails.City
	userProfile.Country = locationDetails.Country
	userProfile.Latitude = locationDetails.Latitude
	userProfile.Longitude = locationDetails.Longitude

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.LocationResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	habits, err := os.getHabits(ctx)
	if err != nil {
		return domain.LocationResult{}, err
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, locationDetails.UserID)
	if err != nil {
		return domain.LocationResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.LocationResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content:         domain.LocationContent{Habits: habits},
	}, nil
}

func (os *onboardingService) Lifestyle(ctx context.Context, lifestyleDetails domain.Lifestyle) (domain.LifestyleResult, error) {
	userProfile, err := os.getUserProfile(ctx, lifestyleDetails.UserID)
	if err != nil {
		return domain.LifestyleResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.MarijuanaID = lifestyleDetails.MarijuanaID
	userProfile.SmokingID = lifestyleDetails.SmokingID
	userProfile.DrugsID = lifestyleDetails.DrugsID
	userProfile.DrinkingID = lifestyleDetails.DrinkingID

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.LifestyleResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	religions, err := os.getReligions(ctx)
	if err != nil {
		return domain.LifestyleResult{}, fmt.Errorf("failed to get religions: %w", err)
	}

	politicalBeliefs, err := os.getPoliticalBeliefs(ctx)
	if err != nil {
		return domain.LifestyleResult{}, fmt.Errorf("failed to get political beliefs: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, lifestyleDetails.UserID)
	if err != nil {
		return domain.LifestyleResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.LifestyleResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.LifestyleContent{
			Religions:        religions,
			PoliticalBeliefs: politicalBeliefs,
		},
	}, nil
}

func (os *onboardingService) Languages(ctx context.Context, spokenLanguages domain.Languages) (domain.LanguagesResult, error) {
	err := os.repo.InsertUserSpokenLanguages(ctx, spokenLanguages.UserID, spokenLanguages.LanguageIDs)
	if err != nil {
		return domain.LanguagesResult{}, fmt.Errorf("failed to insert user spoken languages: %w", err)
	}

	// GET 6 presigned urls from amazon s3
	urls, err := os.awsService.GenerateUploadURLs(ctx, spokenLanguages.UserID, 6, "image/jpeg", time.Duration(10)*time.Minute)
	if err != nil {
		return domain.LanguagesResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var photoUploadUrls []domain.UploadUrl
	for _, url := range urls {
		photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  5242880,
		})
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, spokenLanguages.UserID)
	if err != nil {
		return domain.LanguagesResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.LanguagesResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.LanguagesContent{
			PhotoUploadUrls: photoUploadUrls,
		},
	}, nil
}

func (os *onboardingService) Beliefs(ctx context.Context, beliefDetails domain.Beliefs) (domain.BeliefsResult, error) {
	userProfile, err := os.getUserProfile(ctx, beliefDetails.UserID)
	if err != nil {
		return domain.BeliefsResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.ReligionID = beliefDetails.ReligionID
	userProfile.PoliticalBeliefID = beliefDetails.PoliticalBeliefsID

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.BeliefsResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	educationLevels, err := os.getEducationLevels(ctx)
	if err != nil {
		return domain.BeliefsResult{}, fmt.Errorf("failed to get education levels: %w", err)
	}

	ethnicities, err := os.getEthnicities(ctx)
	if err != nil {
		return domain.BeliefsResult{}, fmt.Errorf("failed to get ethnicities: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, beliefDetails.UserID)
	if err != nil {
		return domain.BeliefsResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.BeliefsResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.BeliefsContent{
			EducationLevels: educationLevels,
			Ethnicities:     ethnicities,
		},
	}, nil
}

func (os *onboardingService) Background(ctx context.Context, backgroundDetails domain.Background) (domain.BackgroundResult, error) {
	userProfile, err := os.getUserProfile(ctx, backgroundDetails.UserID)
	if err != nil {
		return domain.BackgroundResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.EducationLevelID = backgroundDetails.EducationLevelID
	userProfile.EthnicityID = backgroundDetails.EthnicityID

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.BackgroundResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, backgroundDetails.UserID)
	if err != nil {
		return domain.BackgroundResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.BackgroundResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) WorkAndEducation(ctx context.Context, waeDetails domain.WorkAndEducation) (domain.WorkAndEducationResult, error) {
	userProfile, err := os.getUserProfile(ctx, waeDetails.UserID)
	if err != nil {
		return domain.WorkAndEducationResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.Work = &waeDetails.Workplace
	userProfile.JobTitle = &waeDetails.JobTitle
	userProfile.University = &waeDetails.University

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.WorkAndEducationResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	languages, err := os.getLanguages(ctx) // todo: seed languages table
	if err != nil {
		return domain.WorkAndEducationResult{}, fmt.Errorf("failed to get languages: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, waeDetails.UserID)
	if err != nil {
		return domain.WorkAndEducationResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.WorkAndEducationResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.WorkAndEducationContent{
			Languages: languages,
		},
	}, nil
}

func (os *onboardingService) Photos(ctx context.Context, uploadedPhotos domain.UploadedPhotos) (domain.PhotosResult, error) {
	// insert photos into user photos table
	err := os.repo.InsertUserPhotos(ctx, uploadedPhotos.UserID, mapper.MapUploadedPhotosToEntity(uploadedPhotos))
	if err != nil {
		return domain.PhotosResult{}, fmt.Errorf("failed to insert user photos: %w", err)
	}

	// generate prompt urls.
	urls, err := os.awsService.GenerateUploadURLs(ctx, uploadedPhotos.UserID, 6, "audio/m4a", time.Duration(10)*time.Minute)
	if err != nil {
		return domain.PhotosResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var voicePromptUploadUrls []domain.UploadUrl
	for _, url := range urls {
		voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  5242880,
		})
	}

	// get prompts
	prompts, err := os.getPrompts(ctx)
	if err != nil {
		return domain.PhotosResult{}, fmt.Errorf("failed to get prompts: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPhotos.UserID)
	if err != nil {
		return domain.PhotosResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.PhotosResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.PhotosContent{
			Prompts:                prompts,
			VoicePromptsUploadUrls: voicePromptUploadUrls,
		},
	}, nil
}

func (os *onboardingService) Prompts(ctx context.Context, uploadedPrompts domain.Prompts) (domain.PromptsResult, error) {
	// insert prompts into user prompts table
	err := os.repo.InsertUserPrompts(ctx, uploadedPrompts.UserID, mapper.MapPromptsToEntity(uploadedPrompts))
	if err != nil {
		return domain.PromptsResult{}, fmt.Errorf("failed to insert user prompts: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPrompts.UserID)
	if err != nil {
		return domain.PromptsResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.PromptsResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) getLanguages(ctx context.Context) ([]domain.Language, error) {
	languageEntities, err := os.repo.GetLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapLanguagesToDomain(languageEntities), nil
}

func (os *onboardingService) getEducationLevels(ctx context.Context) ([]domain.EducationLevel, error) {
	educationLevelEntities, err := os.repo.GetEducationLevels(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEducationlevelsToDomain(educationLevelEntities), nil
}

func (os *onboardingService) getEthnicities(ctx context.Context) ([]domain.Ethnicity, error) {
	ethnicityEntities, err := os.repo.GetEthnicities(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEthnicityToDomain(ethnicityEntities), nil
}

func (os *onboardingService) getPoliticalBeliefs(ctx context.Context) ([]domain.PoliticalBelief, error) {
	politicalBeliefsEntities, err := os.repo.GetPoliticalBeliefs(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPoliticalBeliefsToDomain(politicalBeliefsEntities), nil
}

func (os *onboardingService) getReligions(ctx context.Context) ([]domain.Religion, error) {
	religionsEntities, err := os.repo.GetReligions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapReligionsToDomain(religionsEntities), nil
}

func (os *onboardingService) getHabits(ctx context.Context) ([]domain.Habit, error) {
	habitEntities, err := os.repo.GetHabits(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapHabitsToDomain(habitEntities), nil
}

func (os *onboardingService) getGenders(ctx context.Context) ([]domain.Gender, error) {
	genderEntities, err := os.repo.GetGenders(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapGendersToDomain(genderEntities), nil
}

func (os *onboardingService) getPrompts(ctx context.Context) ([]domain.Prompt, error) {
	promptEntities, err := os.repo.GetPrompts(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPromptsToDomain(promptEntities), nil
}

func (os *onboardingService) getDatingIntentions(ctx context.Context) ([]domain.DatingIntention, error) {
	datingIntentionsEntities, err := os.repo.GetDatingIntentions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapDatingIntentionsToDomain(datingIntentionsEntities), nil
}

func (os *onboardingService) updateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error {
	updatedUserProfileEntity, err := mapper.MapProfileToEntity(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = os.repo.UpdateUserProfile(ctx, updatedUserProfileEntity)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (os *onboardingService) getUserProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	userProfileEntity, err := os.repo.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapper.MapUserProfileToDomain(userProfileEntity), nil
}

func (os *onboardingService) bumpOnboardingStep(ctx context.Context, userID string) (domain.Steps, error) {
	userDomain, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("failed to get user details: %w", err)
	}

	userDomain.OnboardingStep = string(domain.Steps(userDomain.OnboardingStep).NextStep())

	err = os.userService.UpdateUser(ctx, userDomain)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("failed to update user onboarding step: %w", err)
	}

	return domain.Steps(userDomain.OnboardingStep), nil
}
