package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/theme"
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

const (
	maxUploadCountPhotos  = 6
	maxUploadCountPrompts = 6
	maxUploadBytes        = 5 << 20 // 5 MiB
	presignTTL            = 20 * time.Minute
	mimeJPEG              = "image/jpeg"
	mimeM4A               = "audio/mp4" // m4a is an MP4 container; "audio/m4a" also seen but "audio/mp4" is safer
)

type onboardingService struct {
	logger      *zap.Logger
	profileRepo storage.ProfileRepository
	userService user.Service
	authService auth.Service
	awsService  aws.Service
	lookupRepo  lookupstorage.LookupRepository
}

func NewOnboardingService(
	logger *zap.Logger,
	onboardingRepository storage.ProfileRepository,
	userService user.Service,
	authService auth.Service,
	awsService aws.Service,
	lookupRepo lookupstorage.LookupRepository,
) Service {
	return &onboardingService{
		logger:      logger,
		profileRepo: onboardingRepository,
		userService: userService,
		authService: authService,
		awsService:  awsService,
		lookupRepo:  lookupRepo,
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
		//todo: update to depend on the media service instead
		urls, err := os.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPhotos, mimeJPEG, presignTTL)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
		}

		var photoUploadUrls []domain.UploadUrl
		for _, url := range urls {
			photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
				Key:       url.Key,
				UploadUrl: url.URL,
				Headers:   url.Headers,
				MaxBytes:  maxUploadBytes,
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
		urls, err := os.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPrompts, mimeM4A, presignTTL)
		if err != nil {
			return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
		}

		var voicePromptUploadUrls []domain.UploadUrl
		for _, url := range urls {
			voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
				Key:       url.Key,
				UploadUrl: url.URL,
				Headers:   url.Headers,
				MaxBytes:  maxUploadBytes,
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

	userProfile, err := os.getUserProfile(ctx, introDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	var Lastname string
	if introDetails.LastName != nil {
		Lastname = *introDetails.LastName
	}

	displayName := fmt.Sprintf("%s %s", introDetails.FirstName, Lastname)
	userProfile.DisplayName = &displayName

	err = os.updateUserProfile(ctx, userProfile)
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

	userProfile, err := os.getUserProfile(ctx, basicDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.GenderID = basicDetails.GenderID
	userProfile.DatingIntentionID = basicDetails.DatingIntentionID
	userProfile.HeightCM = basicDetails.HeightCm
	userProfile.Birthdate = basicDetails.Birthdate

	err = os.updateUserProfile(ctx, userProfile)
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

	userProfile, err := os.getUserProfile(ctx, locationDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.City = locationDetails.City
	userProfile.Country = locationDetails.Country
	userProfile.Latitude = locationDetails.Latitude
	userProfile.Longitude = locationDetails.Longitude

	err = os.updateUserProfile(ctx, userProfile)
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

	userProfile, err := os.getUserProfile(ctx, lifestyleDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.MarijuanaID = lifestyleDetails.MarijuanaID
	userProfile.SmokingID = lifestyleDetails.SmokingID
	userProfile.DrugsID = lifestyleDetails.DrugsID
	userProfile.DrinkingID = lifestyleDetails.DrinkingID

	err = os.updateUserProfile(ctx, userProfile)
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

	err = os.profileRepo.UpsertUserSpokenLanguages(ctx, spokenLanguages.UserID, spokenLanguages.LanguageIDs)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to insert user spoken languages: %w", err)
	}

	// GET 6 presigned urls from amazon s3
	urls, err := os.awsService.GenerateUploadURLs(ctx, spokenLanguages.UserID, maxUploadCountPhotos, mimeJPEG, presignTTL)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var photoUploadUrls []domain.UploadUrl
	for _, url := range urls {
		photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
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

	userProfile, err := os.getUserProfile(ctx, beliefDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.ReligionID = beliefDetails.ReligionID
	userProfile.PoliticalBeliefID = beliefDetails.PoliticalBeliefsID

	err = os.updateUserProfile(ctx, userProfile)
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

	userProfile, err := os.getUserProfile(ctx, backgroundDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.EducationLevelID = backgroundDetails.EducationLevelID
	userProfile.EthnicityID = backgroundDetails.EthnicityID

	err = os.updateUserProfile(ctx, userProfile)
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

	userProfile, err := os.getUserProfile(ctx, waeDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.Work = &waeDetails.Workplace
	userProfile.JobTitle = &waeDetails.JobTitle
	userProfile.University = &waeDetails.University

	err = os.updateUserProfile(ctx, userProfile)
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
	err = os.profileRepo.UpsertUserPhotos(ctx, uploadedPhotos.UserID, mapper.MapUploadedPhotosToEntity(uploadedPhotos))
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to insert user photos: %w", err)
	}

	// generate prompt urls.
	urls, err := os.awsService.GenerateUploadURLs(ctx, uploadedPhotos.UserID, maxUploadCountPrompts, mimeM4A, presignTTL)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to generate upload urls: %w", err)
	}

	var voicePromptUploadUrls []domain.UploadUrl
	for _, url := range urls {
		voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
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

	// insert prompts into user prompts table
	err = os.profileRepo.UpsertUserPrompts(ctx, uploadedPrompts.UserID, mapper.MapPromptsToEntity(uploadedPrompts))
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

	userProfile, err := os.getUserProfile(ctx, profileDetails.UserID)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile.CoverPhotoURL = &profileDetails.ProfileCoverPhotoURL

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	// generate colours
	palette, err := theme.GeneratePalette9(profileDetails.ProfileBaseColour)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to generate palette: %w", err)
	}

	palJSON, err := json.Marshal(palette)
	if err != nil {
		return domain.StepResult{}, fmt.Errorf("failed to marshal palette: %w", err)
	}
	// store colours.
	err = os.profileRepo.UpsertUserTheme(ctx, entity.UserTheme{
		UserID:  profileDetails.UserID,
		BaseHex: profileDetails.ProfileBaseColour,
		Palette: palJSON,
	})
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

func (os *onboardingService) getLanguages(ctx context.Context) ([]domain.Language, error) {
	languageEntities, err := os.lookupRepo.GetLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapLanguagesToDomain(languageEntities), nil
}

func (os *onboardingService) getEducationLevels(ctx context.Context) ([]domain.EducationLevel, error) {
	educationLevelEntities, err := os.lookupRepo.GetEducationLevels(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEducationlevelsToDomain(educationLevelEntities), nil
}

func (os *onboardingService) getEthnicities(ctx context.Context) ([]domain.Ethnicity, error) {
	ethnicityEntities, err := os.lookupRepo.GetEthnicities(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEthnicityToDomain(ethnicityEntities), nil
}

func (os *onboardingService) getPoliticalBeliefs(ctx context.Context) ([]domain.PoliticalBelief, error) {
	politicalBeliefsEntities, err := os.lookupRepo.GetPoliticalBeliefs(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPoliticalBeliefsToDomain(politicalBeliefsEntities), nil
}

func (os *onboardingService) getReligions(ctx context.Context) ([]domain.Religion, error) {
	religionsEntities, err := os.lookupRepo.GetReligions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapReligionsToDomain(religionsEntities), nil
}

func (os *onboardingService) getHabits(ctx context.Context) ([]domain.Habit, error) {
	habitEntities, err := os.lookupRepo.GetHabits(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapHabitsToDomain(habitEntities), nil
}

func (os *onboardingService) getGenders(ctx context.Context) ([]domain.Gender, error) {
	genderEntities, err := os.lookupRepo.GetGenders(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapGendersToDomain(genderEntities), nil
}

func (os *onboardingService) getPrompts(ctx context.Context) ([]domain.Prompt, error) {
	promptEntities, err := os.lookupRepo.GetPrompts(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPromptsToDomain(promptEntities), nil
}

func (os *onboardingService) getDatingIntentions(ctx context.Context) ([]domain.DatingIntention, error) {
	datingIntentionsEntities, err := os.lookupRepo.GetDatingIntentions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapDatingIntentionsToDomain(datingIntentionsEntities), nil
}

func (os *onboardingService) updateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error {
	updatedUserProfileEntity, whiteList, err := mapper.MapProfileToEntityForUpdate(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = os.profileRepo.UpdateUserProfile(ctx, updatedUserProfileEntity, whiteList)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (os *onboardingService) getUserProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	userProfileEntity, err := os.profileRepo.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapper.MapUserProfileToDomain(userProfileEntity), nil
}

// ensureStep makes sure that the step being called is the correct step to complete for the provided user. This prevents user's skipping steps in the onboarding process.
func (os *onboardingService) ensureStep(ctx context.Context, userID string, expected domain.Steps) error {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if domain.Steps(u.OnboardingStep) != expected {
		return ErrIncorrectStepCalled
	}

	return nil
}

func (os *onboardingService) bumpOnboardingStep(
	ctx context.Context,
	userID string,
	expected domain.Steps, // the step this endpoint owns
) (domain.Steps, error) {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("get user: %w", err)
	}

	curr := domain.Steps(u.OnboardingStep)

	// 1) If we’re not exactly at the expected step, do NOT advance.
	//    (Already advanced or out of order => no-op)
	if curr != expected {
		return curr, nil
	}

	next := expected.NextStep()
	u.OnboardingStep = string(next)

	err = os.userService.UpdateUser(ctx, u)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("update step: %w", err)
	}

	return next, nil
}
