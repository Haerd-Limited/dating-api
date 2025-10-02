package onboarding

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/media"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

type Service interface {
	// GetUserCurrentStep retrieves the current onboarding step for a user based on their unique user ID. Returns a GetUserCurrentStepResult.
	GetUserCurrentStep(ctx context.Context, userID string) (domain.StepResult, error)
	// Intro creates a new user based on the provided registration details and returns tokens and onboarding steps.
	Intro(ctx context.Context, introDetails domain.Intro) (domain.StepResult, error)
	// Basics updates the user's basic details such as birthdate, height, gender, and dating intention, and returns onboarding steps.
	Basics(ctx context.Context, basicDetails domain.Basics) (domain.StepResult, error)
	// Location updates the user's location details and returns the result of the operation or an error if it fails.
	Location(ctx context.Context, locationDetails domain.Location) (domain.StepResult, error)
	// Lifestyle updates the user's lifestyle preferences and returns the updated onboarding steps and lifestyle content.
	Lifestyle(ctx context.Context, lifestyleDetails domain.Lifestyle) (domain.StepResult, error)
	// Beliefs processes and updates the user's political and religious beliefs details during onboarding.
	Beliefs(ctx context.Context, beliefDetails domain.Beliefs) (domain.StepResult, error)
	// Background updates the user's background details and returns the result of the operation or an error if any occurs.
	Background(ctx context.Context, backgroundDetails domain.Background) (domain.StepResult, error)
	// WorkAndEducation handles the onboarding step to update work and education data for a user.
	WorkAndEducation(ctx context.Context, waeDetails domain.WorkAndEducation) (domain.StepResult, error)
	// Languages updates the user's spoken languages and returns the updated onboarding steps and language content.
	Languages(ctx context.Context, spokenLanguages domain.Languages) (domain.StepResult, error)
	// Photos handles the processing and storage of uploaded photos for a user and updates their onboarding progress.
	Photos(ctx context.Context, uploadedPhotos domain.UploadedPhotos) (domain.StepResult, error)
	// Prompts handles the submission of user-uploaded prompts and updates the onboarding process with the provided data.
	Prompts(ctx context.Context, uploadedPrompts domain.Prompts) (domain.StepResult, error)
	Profile(ctx context.Context, profileDetails domain.Profile) (domain.StepResult, error)
}

type onboardingService struct {
	logger         *zap.Logger
	userService    user.Service
	authService    auth.Service
	lookupRepo     lookupstorage.LookupRepository
	mediaService   media.Service
	profileService profile.Service
}

func NewOnboardingService(
	logger *zap.Logger,
	userService user.Service,
	authService auth.Service,
	mediaService media.Service,
	profileService profile.Service,
	lookupRepo lookupstorage.LookupRepository,
) Service {
	return &onboardingService{
		logger:         logger,
		userService:    userService,
		authService:    authService,
		lookupRepo:     lookupRepo,
		mediaService:   mediaService,
		profileService: profileService,
	}
}

var ErrIncorrectStepCalled = errors.New("incorrect step called")

func (os *onboardingService) GetUserCurrentStep(ctx context.Context, userID string) (domain.StepResult, error) {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user by userID: %w", err)
	}

	currentStep := domain.Steps(u.OnboardingStep)

	// each step should return the previous step's response since a user being on a step means they've not completed that step.
	switch currentStep {
	case domain.OnboardingStepsIntro:
		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
		}, nil
	case domain.OnboardingStepsBasics:
		genders, err := os.getGenders(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get genders: %w", err)
		}

		datingIntentions, err := os.getDatingIntentions(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get dating intentions: %w", err)
		}

		return domain.StepResult{
			OnboardingSteps: domain.OnboardingStepsBasics.GenerateOnboardingSteps(),
			Content: domain.IntroContent{
				DatingIntentions: datingIntentions,
				Genders:          genders,
			},
		}, nil
	case domain.OnboardingStepsLocation:
		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
		}, nil
	case domain.OnboardingStepsLifestyle:
		habits, err := os.getHabits(ctx)
		if err != nil {
			return domain.StepResult{}, err
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content:         domain.LocationContent{Habits: habits},
		}, nil

	case domain.OnboardingStepsBeliefs:
		religions, err := os.getReligions(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get religions: %w", err)
		}

		politicalBeliefs, err := os.getPoliticalBeliefs(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get political beliefs: %w", err)
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content: domain.LifestyleContent{
				Religions:        religions,
				PoliticalBeliefs: politicalBeliefs,
			},
		}, nil
	case domain.OnboardingStepsBackground:
		educationLevels, err := os.getEducationLevels(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get education levels: %w", err)
		}

		ethnicities, err := os.getEthnicities(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get ethnicities: %w", err)
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content: domain.BeliefsContent{
				EducationLevels: educationLevels,
				Ethnicities:     ethnicities,
			},
		}, nil

	case domain.OnboardingStepsWorkAndEducation:
		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
		}, nil
	case domain.OnboardingStepsLanguages:
		languages, err := os.getLanguages(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get languages: %w", err)
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content: domain.WorkAndEducationContent{
				Languages: languages,
			},
		}, nil

	case domain.OnboardingStepsPhotos:
		// GET 6 presigned urls from amazon s3
		urls, err := os.mediaService.GenerateUploadURLsForProfilePhotos(ctx, userID)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to generate profile photo upload urls: %w", err)
		}

		var photoUploadUrls []domain.UploadUrl
		for _, url := range urls {
			photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
				Key:       url.Key,
				UploadUrl: url.UploadUrl,
				Headers:   url.Headers,
				MaxBytes:  url.MaxBytes,
			})
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content: domain.LanguagesContent{
				PhotoUploadUrls: photoUploadUrls,
			},
		}, nil

	case domain.OnboardingStepsPrompts:
		// generate prompt urls.
		urls, err := os.mediaService.GenerateUploadURLsForProfilePrompts(ctx, userID)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
		}

		var voicePromptUploadUrls []domain.UploadUrl
		for _, url := range urls {
			voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
				Key:       url.Key,
				UploadUrl: url.UploadUrl,
				Headers:   url.Headers,
				MaxBytes:  url.MaxBytes,
			})
		}

		// get prompts
		prompts, err := os.getPrompts(ctx)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to get prompts: %w", err)
		}

		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
			Content: domain.PhotosContent{
				Prompts:                prompts,
				VoicePromptsUploadUrls: voicePromptUploadUrls,
			},
		}, nil
	case domain.OnboardingStepsProfile:
		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
		}, nil

	case domain.OnboardingStepsComplete:
		return domain.StepResult{
			OnboardingSteps: currentStep.GenerateOnboardingSteps(),
		}, nil
	default:
		return domain.StepResult{}, fmt.Errorf("unknown onboarding step: %s", currentStep)
	}
}

func (os *onboardingService) Intro(ctx context.Context, introDetails domain.Intro) (domain.StepResult, error) {
	const StepForIntro = domain.OnboardingStepsIntro

	err := os.ensureStep(ctx, introDetails.UserID, StepForIntro)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	err = os.userService.UpdateUser(ctx, &userdomain.User{
		ID:        introDetails.UserID,
		Email:     introDetails.Email,
		FirstName: introDetails.FirstName,
		LastName:  introDetails.LastName,
	})
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, introDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	var Lastname string
	if introDetails.LastName != nil {
		Lastname = *introDetails.LastName
	}

	displayName := fmt.Sprintf("%s %s", introDetails.FirstName, Lastname)
	userProfile.DisplayName = &displayName

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	// Get dating intentions and genders and populate Content
	genders, err := os.getGenders(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get genders: %w", err)
	}

	datingIntentions, err := os.getDatingIntentions(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get dating intentions: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, introDetails.UserID, StepForIntro)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.IntroContent{
			DatingIntentions: datingIntentions,
			Genders:          genders,
		},
	}, nil
}

func (os *onboardingService) Basics(ctx context.Context, basicDetails domain.Basics) (domain.StepResult, error) {
	const StepForBasics = domain.OnboardingStepsBasics

	err := os.ensureStep(ctx, basicDetails.UserID, StepForBasics)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, basicDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.GenderID = &basicDetails.GenderID
	userProfile.DatingIntentionID = &basicDetails.DatingIntentionID
	userProfile.HeightCM = &basicDetails.HeightCm
	userProfile.Birthdate = &basicDetails.Birthdate

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, basicDetails.UserID, StepForBasics)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) Location(ctx context.Context, locationDetails domain.Location) (domain.StepResult, error) {
	const StepForLocation = domain.OnboardingStepsLocation

	err := os.ensureStep(ctx, locationDetails.UserID, StepForLocation)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, locationDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.City = &locationDetails.City
	userProfile.Country = &locationDetails.Country
	userProfile.Latitude = &locationDetails.Latitude
	userProfile.Longitude = &locationDetails.Longitude

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	habits, err := os.getHabits(ctx)
	if err != nil {
		return domain.StepResult{}, err
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, locationDetails.UserID, StepForLocation)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content:         domain.LocationContent{Habits: habits},
	}, nil
}

func (os *onboardingService) Lifestyle(ctx context.Context, lifestyleDetails domain.Lifestyle) (domain.StepResult, error) {
	const StepForLifestyle = domain.OnboardingStepsLifestyle

	err := os.ensureStep(ctx, lifestyleDetails.UserID, StepForLifestyle)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, lifestyleDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.MarijuanaID = &lifestyleDetails.MarijuanaID
	userProfile.SmokingID = &lifestyleDetails.SmokingID
	userProfile.DrugsID = &lifestyleDetails.DrugsID
	userProfile.DrinkingID = &lifestyleDetails.DrinkingID

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	religions, err := os.getReligions(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get religions: %w", err)
	}

	politicalBeliefs, err := os.getPoliticalBeliefs(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get political beliefs: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, lifestyleDetails.UserID, StepForLifestyle)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.LifestyleContent{
			Religions:        religions,
			PoliticalBeliefs: politicalBeliefs,
		},
	}, nil
}

func (os *onboardingService) Languages(ctx context.Context, spokenLanguages domain.Languages) (domain.StepResult, error) {
	const StepForLanguages = domain.OnboardingStepsLanguages

	err := os.ensureStep(ctx, spokenLanguages.UserID, StepForLanguages)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	err = os.profileService.UpsertUserSpokenLanguages(ctx, spokenLanguages.UserID, spokenLanguages.LanguageIDs)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to insert user spoken languages: %w", err)
	}

	// GET 6 presigned urls from amazon s3
	urls, err := os.mediaService.GenerateUploadURLsForProfilePhotos(ctx, spokenLanguages.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var photoUploadUrls []domain.UploadUrl
	for _, url := range urls {
		photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.UploadUrl,
			Headers:   url.Headers,
			MaxBytes:  url.MaxBytes,
		})
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, spokenLanguages.UserID, StepForLanguages)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.LanguagesContent{
			PhotoUploadUrls: photoUploadUrls,
		},
	}, nil
}

func (os *onboardingService) Beliefs(ctx context.Context, beliefDetails domain.Beliefs) (domain.StepResult, error) {
	const StepForBeliefs = domain.OnboardingStepsBeliefs

	err := os.ensureStep(ctx, beliefDetails.UserID, StepForBeliefs)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, beliefDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.ReligionID = &beliefDetails.ReligionID
	userProfile.PoliticalBeliefID = &beliefDetails.PoliticalBeliefsID

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	educationLevels, err := os.getEducationLevels(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get education levels: %w", err)
	}

	ethnicities, err := os.getEthnicities(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get ethnicities: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, beliefDetails.UserID, StepForBeliefs)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.BeliefsContent{
			EducationLevels: educationLevels,
			Ethnicities:     ethnicities,
		},
	}, nil
}

func (os *onboardingService) Background(ctx context.Context, backgroundDetails domain.Background) (domain.StepResult, error) {
	const StepForBackground = domain.OnboardingStepsBackground

	err := os.ensureStep(ctx, backgroundDetails.UserID, StepForBackground)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, backgroundDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.EducationLevelID = &backgroundDetails.EducationLevelID
	userProfile.EthnicityID = &backgroundDetails.EthnicityID

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, backgroundDetails.UserID, StepForBackground)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) WorkAndEducation(ctx context.Context, waeDetails domain.WorkAndEducation) (domain.StepResult, error) {
	const StepForWorkAndEducation = domain.OnboardingStepsWorkAndEducation

	err := os.ensureStep(ctx, waeDetails.UserID, StepForWorkAndEducation)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, waeDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.Work = &waeDetails.Workplace
	userProfile.JobTitle = &waeDetails.JobTitle
	userProfile.University = &waeDetails.University

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	languages, err := os.getLanguages(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get languages: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, waeDetails.UserID, StepForWorkAndEducation)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.WorkAndEducationContent{
			Languages: languages,
		},
	}, nil
}

func (os *onboardingService) Photos(ctx context.Context, uploadedPhotos domain.UploadedPhotos) (domain.StepResult, error) {
	const StepForPhotos = domain.OnboardingStepsPhotos

	err := os.ensureStep(ctx, uploadedPhotos.UserID, StepForPhotos)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	// insert photos into user photos table
	err = os.profileService.UpsertUserPhotos(ctx, uploadedPhotos.UserID, mapper.MapUploadedPhotosToProfilePhotos(uploadedPhotos))
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to insert user photos: %w", err)
	}

	// generate prompt urls.
	urls, err := os.mediaService.GenerateUploadURLsForProfilePrompts(ctx, uploadedPhotos.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var voicePromptUploadUrls []domain.UploadUrl
	for _, url := range urls {
		voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.UploadUrl,
			Headers:   url.Headers,
			MaxBytes:  url.MaxBytes,
		})
	}

	// get prompts
	prompts, err := os.getPrompts(ctx)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get prompts: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPhotos.UserID, StepForPhotos)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content: domain.PhotosContent{
			Prompts:                prompts,
			VoicePromptsUploadUrls: voicePromptUploadUrls,
		},
	}, nil
}

func (os *onboardingService) Prompts(ctx context.Context, uploadedPrompts domain.Prompts) (domain.StepResult, error) {
	const StepForPrompts = domain.OnboardingStepsPrompts

	err := os.ensureStep(ctx, uploadedPrompts.UserID, StepForPrompts)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	// todo: ensure count is min 4
	// insert prompts into user prompts table
	err = os.profileService.UpsertUserPrompts(ctx, uploadedPrompts.UserID, mapper.MapPromptsToProfileVoicePrompts(uploadedPrompts))
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to insert user prompts: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPrompts.UserID, StepForPrompts)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) Profile(ctx context.Context, profileDetails domain.Profile) (domain.StepResult, error) {
	const StepForProfile = domain.OnboardingStepsProfile

	err := os.ensureStep(ctx, profileDetails.UserID, StepForProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to ensure step: %w", err)
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, profileDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.CoverPhotoURL = &profileDetails.ProfileCoverPhotoURL

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	err = os.profileService.UpsertUserTheme(ctx, profileDetails.UserID, profileDetails.ProfileBaseColour)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to upsert user theme: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, profileDetails.UserID, StepForProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}
