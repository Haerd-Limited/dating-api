package onboarding

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/media"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	// GetUserCurrentStep retrieves the current onboarding step for a user based on their unique user ID. Returns a GetUserCurrentStepResult.
	GetUserCurrentStep(ctx context.Context, userID string) (domain.StepResult, error)
	// Intro creates a new user based on the provided registration details and returns tokens and onboarding steps.
	Intro(ctx context.Context, introDetails domain.Intro) (domain.StepResult, error)
	// Basics updates the user's basic details such as birthdate, height, gender, and dating intention, and returns onboarding steps.
	Basics(ctx context.Context, basicDetails domain.Basics) (domain.StepResult, error)
	// GetPreregistrationStats returns current counts and configured limits for the landing page.
	GetPreregistrationStats(ctx context.Context) (domain.PreregistrationStats, error)
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
	// prereg caps
	enablePreregCap       bool
	maxTotalParticipants  int
	maxMaleParticipants   int
	maxFemaleParticipants int
}

func NewOnboardingService(
	logger *zap.Logger,
	userService user.Service,
	authService auth.Service,
	mediaService media.Service,
	profileService profile.Service,
	lookupRepo lookupstorage.LookupRepository,
	enablePreregCap bool,
	maxTotal int,
	maxMale int,
	maxFemale int,
) Service {
	return &onboardingService{
		logger:                logger,
		userService:           userService,
		authService:           authService,
		lookupRepo:            lookupRepo,
		mediaService:          mediaService,
		profileService:        profileService,
		enablePreregCap:       enablePreregCap,
		maxTotalParticipants:  maxTotal,
		maxMaleParticipants:   maxMale,
		maxFemaleParticipants: maxFemale,
	}
}

var (
	ErrIncorrectStepCalled      = errors.New("incorrect step called")
	ErrMissingPrompts           = errors.New("missing prompts")
	ErrNotEnoughPromptsProvided = errors.New("not enough prompts provided")
	ErrTooManyPromptsProvided   = errors.New("too many prompts provided")
	ErrPreregistrationCapped    = errors.New("preregistration cap reached")
)

const (
	MinimumNumberOfPrompts = 4
)

func (os *onboardingService) GetUserCurrentStep(ctx context.Context, userID string) (domain.StepResult, error) {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user by userID", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get genders", err, zap.String("userID", userID))
		}

		datingIntentions, err := os.getDatingIntentions(ctx)
		if err != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get dating intentions", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get religions", err, zap.String("userID", userID))
		}

		politicalBeliefs, err := os.getPoliticalBeliefs(ctx)
		if err != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get political beliefs", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get education levels", err, zap.String("userID", userID))
		}

		ethnicities, err := os.getEthnicities(ctx)
		if err != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get ethnicities", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get languages", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "generate profile photo upload urls", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "generate upload urls", err, zap.String("userID", userID))
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
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get prompts", err, zap.String("userID", userID))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "unknown onboarding step", errors.New("unknown step"), zap.String("userID", userID), zap.String("step", string(currentStep)))
	}
}

func (os *onboardingService) Intro(ctx context.Context, introDetails domain.Intro) (domain.StepResult, error) {
	const StepForIntro = domain.OnboardingStepsIntro

	err := os.ensureStep(ctx, introDetails.UserID, StepForIntro)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", introDetails.UserID), zap.String("step", string(StepForIntro)))
	}

	err = os.userService.UpdateUser(ctx, &userdomain.User{
		ID:        introDetails.UserID,
		Email:     introDetails.Email,
		FirstName: introDetails.FirstName,
		LastName:  introDetails.LastName,
	})
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user", err, zap.String("userID", introDetails.UserID))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, introDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", introDetails.UserID))
	}

	var Lastname string
	if introDetails.LastName != nil {
		Lastname = *introDetails.LastName
	}

	displayName := fmt.Sprintf("%s %s", introDetails.FirstName, Lastname)
	userProfile.DisplayName = &displayName

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", introDetails.UserID), zap.Any("userProfile", userProfile))
	}

	// Get dating intentions and genders and populate Content
	genders, err := os.getGenders(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get genders", err, zap.String("userID", introDetails.UserID))
	}

	datingIntentions, err := os.getDatingIntentions(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get dating intentions", err, zap.String("userID", introDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, introDetails.UserID, StepForIntro)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", introDetails.UserID), zap.String("step", string(StepForIntro)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", basicDetails.UserID), zap.String("step", string(StepForBasics)))
	}

	// Enforce prereg caps at the moment a user tries to complete BASICS (which moves them to LOCATION).
	if os.enablePreregCap {
		genderEntity, gErr := os.lookupRepo.GetGenderByID(ctx, basicDetails.GenderID)
		if gErr != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "get gender", gErr, zap.String("userID", basicDetails.UserID), zap.Int16("genderID", basicDetails.GenderID))
		}

		totalCompleted, tErr := os.profileService.CountBasicsCompleted(ctx)
		if tErr != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "count basics-completed total", tErr, zap.String("userID", basicDetails.UserID))
		}

		if os.maxTotalParticipants > 0 && int(totalCompleted) >= os.maxTotalParticipants {
			if delErr := os.userService.DeleteAccount(ctx, basicDetails.UserID); delErr != nil {
				return domain.StepResult{}, commonlogger.LogError(os.logger, "delete account after total cap reached", delErr, zap.String("userID", basicDetails.UserID))
			}

			return domain.StepResult{}, ErrPreregistrationCapped
		}

		byGender, cErr := os.profileService.CountBasicsCompletedByGender(ctx, basicDetails.GenderID)
		if cErr != nil {
			return domain.StepResult{}, commonlogger.LogError(os.logger, "count basics-completed by gender", cErr, zap.String("userID", basicDetails.UserID), zap.Int16("genderID", basicDetails.GenderID))
		}

		switch genderEntity.Label {
		case "Male":
			if os.maxMaleParticipants > 0 && int(byGender) >= os.maxMaleParticipants {
				if delErr := os.userService.DeleteAccount(ctx, basicDetails.UserID); delErr != nil {
					return domain.StepResult{}, commonlogger.LogError(os.logger, "delete account after male cap reached", delErr, zap.String("userID", basicDetails.UserID))
				}

				return domain.StepResult{}, ErrPreregistrationCapped
			}
		case "Female":
			if os.maxFemaleParticipants > 0 && int(byGender) >= os.maxFemaleParticipants {
				if delErr := os.userService.DeleteAccount(ctx, basicDetails.UserID); delErr != nil {
					return domain.StepResult{}, commonlogger.LogError(os.logger, "delete account after female cap reached", delErr, zap.String("userID", basicDetails.UserID))
				}

				return domain.StepResult{}, ErrPreregistrationCapped
			}
		}
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, basicDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", basicDetails.UserID))
	}

	userProfile.GenderID = &basicDetails.GenderID
	userProfile.DatingIntentionID = &basicDetails.DatingIntentionID
	userProfile.HeightCM = &basicDetails.HeightCm

	dob, err := time.Parse(time.DateOnly, basicDetails.Birthdate)
	if err != nil {
		return domain.StepResult{}, commonErrors.ErrInvalidDob
	}

	userProfile.Birthdate = &dob

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", basicDetails.UserID), zap.Any("userProfile", userProfile))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, basicDetails.UserID, StepForBasics)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", basicDetails.UserID), zap.String("step", string(StepForBasics)))
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) Location(ctx context.Context, locationDetails domain.Location) (domain.StepResult, error) {
	const StepForLocation = domain.OnboardingStepsLocation

	err := os.ensureStep(ctx, locationDetails.UserID, StepForLocation)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", locationDetails.UserID), zap.String("step", string(StepForLocation)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, locationDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", locationDetails.UserID))
	}

	userProfile.City = &locationDetails.City
	userProfile.Country = &locationDetails.Country
	userProfile.Latitude = &locationDetails.Latitude
	userProfile.Longitude = &locationDetails.Longitude

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", locationDetails.UserID), zap.Any("userProfile", userProfile))
	}

	habits, err := os.getHabits(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get habits", err, zap.String("userID", locationDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, locationDetails.UserID, StepForLocation)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", locationDetails.UserID), zap.String("step", string(StepForLocation)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", lifestyleDetails.UserID), zap.String("step", string(StepForLifestyle)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, lifestyleDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", lifestyleDetails.UserID))
	}

	userProfile.MarijuanaID = &lifestyleDetails.MarijuanaID
	userProfile.SmokingID = &lifestyleDetails.SmokingID
	userProfile.DrugsID = &lifestyleDetails.DrugsID
	userProfile.DrinkingID = &lifestyleDetails.DrinkingID

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", lifestyleDetails.UserID), zap.Any("userProfile", userProfile))
	}

	religions, err := os.getReligions(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get religions", err, zap.String("userID", lifestyleDetails.UserID))
	}

	politicalBeliefs, err := os.getPoliticalBeliefs(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get political beliefs", err, zap.String("userID", lifestyleDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, lifestyleDetails.UserID, StepForLifestyle)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", lifestyleDetails.UserID), zap.String("step", string(StepForLifestyle)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", spokenLanguages.UserID), zap.String("step", string(StepForLanguages)))
	}

	err = os.profileService.UpsertUserSpokenLanguages(ctx, spokenLanguages.UserID, spokenLanguages.LanguageIDs)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "insert user spoken languages", err, zap.String("userID", spokenLanguages.UserID))
	}

	// GET 6 presigned urls from amazon s3
	urls, err := os.mediaService.GenerateUploadURLsForProfilePhotos(ctx, spokenLanguages.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "generate upload urls", err, zap.String("userID", spokenLanguages.UserID))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", spokenLanguages.UserID), zap.String("step", string(StepForLanguages)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", beliefDetails.UserID), zap.String("step", string(StepForBeliefs)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, beliefDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", beliefDetails.UserID))
	}

	userProfile.ReligionID = &beliefDetails.ReligionID
	userProfile.PoliticalBeliefID = &beliefDetails.PoliticalBeliefsID

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", beliefDetails.UserID), zap.Any("userProfile", userProfile))
	}

	educationLevels, err := os.getEducationLevels(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get education levels", err, zap.String("userID", beliefDetails.UserID))
	}

	ethnicities, err := os.getEthnicities(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get ethnicities", err, zap.String("userID", beliefDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, beliefDetails.UserID, StepForBeliefs)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", beliefDetails.UserID), zap.String("step", string(StepForBeliefs)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", backgroundDetails.UserID), zap.String("step", string(StepForBackground)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, backgroundDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", backgroundDetails.UserID))
	}

	userProfile.EducationLevelID = &backgroundDetails.EducationLevelID
	userProfile.EthnicityIDs = backgroundDetails.EthnicityIDs

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", backgroundDetails.UserID), zap.Any("userProfile", userProfile))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, backgroundDetails.UserID, StepForBackground)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", backgroundDetails.UserID), zap.String("step", string(StepForBackground)))
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) WorkAndEducation(ctx context.Context, waeDetails domain.WorkAndEducation) (domain.StepResult, error) {
	const StepForWorkAndEducation = domain.OnboardingStepsWorkAndEducation

	err := os.ensureStep(ctx, waeDetails.UserID, StepForWorkAndEducation)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", waeDetails.UserID), zap.String("step", string(StepForWorkAndEducation)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, waeDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", waeDetails.UserID))
	}

	userProfile.Work = &waeDetails.Workplace
	userProfile.JobTitle = &waeDetails.JobTitle
	userProfile.University = &waeDetails.University

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", waeDetails.UserID), zap.Any("userProfile", userProfile))
	}

	languages, err := os.getLanguages(ctx)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get languages", err, zap.String("userID", waeDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, waeDetails.UserID, StepForWorkAndEducation)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", waeDetails.UserID), zap.String("step", string(StepForWorkAndEducation)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", uploadedPhotos.UserID), zap.String("step", string(StepForPhotos)))
	}

	// insert photos into user photos table
	err = os.profileService.UpsertUserPhotos(ctx, uploadedPhotos.UserID, mapper.MapUploadedPhotosToProfilePhotos(uploadedPhotos))
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "insert user photos", err, zap.String("userID", uploadedPhotos.UserID), zap.Any("uploadedPhotos", uploadedPhotos))
	}

	// generate prompt urls.
	urls, err := os.mediaService.GenerateUploadURLsForProfilePrompts(ctx, uploadedPhotos.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "generate upload urls", err, zap.String("userID", uploadedPhotos.UserID))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get prompts", err, zap.String("userID", uploadedPhotos.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPhotos.UserID, StepForPhotos)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", uploadedPhotos.UserID), zap.String("step", string(StepForPhotos)))
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
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", uploadedPrompts.UserID), zap.String("step", string(StepForPrompts)))
	}

	err = os.validatePrompts(uploadedPrompts)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "validate prompts", err, zap.String("userID", uploadedPrompts.UserID))
	}

	// insert prompts into user prompts table
	err = os.profileService.UpsertUserPrompts(ctx, uploadedPrompts.UserID, mapper.MapPromptsToProfileVoicePrompts(uploadedPrompts))
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "insert user prompts", err, zap.String("userID", uploadedPrompts.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, uploadedPrompts.UserID, StepForPrompts)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", uploadedPrompts.UserID), zap.String("step", string(StepForPrompts)))
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) Profile(ctx context.Context, profileDetails domain.Profile) (domain.StepResult, error) {
	const StepForProfile = domain.OnboardingStepsProfile

	err := os.ensureStep(ctx, profileDetails.UserID, StepForProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "ensure step", err, zap.String("userID", profileDetails.UserID), zap.String("step", string(StepForProfile)))
	}

	userProfile, err := os.profileService.GetProfileForUpdate(ctx, profileDetails.UserID)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "get user profile by userID", err, zap.String("userID", profileDetails.UserID))
	}

	userProfile.CoverPhotoURL = &profileDetails.ProfileCoverPhotoURL

	err = os.profileService.UpdateProfile(ctx, userProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "update user profile", err, zap.String("userID", profileDetails.UserID), zap.Any("userProfile", userProfile))
	}

	err = os.profileService.UpsertUserTheme(ctx, profileDetails.UserID, profileDetails.ProfileBaseColour)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "upsert user theme", err, zap.String("userID", profileDetails.UserID))
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, profileDetails.UserID, StepForProfile)
	if err != nil {
		return domain.StepResult{}, commonlogger.LogError(os.logger, "bump onboarding step", err, zap.String("userID", profileDetails.UserID), zap.String("step", string(StepForProfile)))
	}

	return domain.StepResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
	}, nil
}

func (os *onboardingService) GetPreregistrationStats(ctx context.Context) (domain.PreregistrationStats, error) {
	// Resolve gender IDs for Male/Female
	genders, err := os.lookupRepo.GetGenders(ctx)
	if err != nil {
		return domain.PreregistrationStats{}, commonlogger.LogError(os.logger, "get genders", err)
	}

	var maleID int16

	var femaleID int16

	for _, g := range genders {
		switch g.Label {
		case "Male":
			maleID = g.ID
		case "Female":
			femaleID = g.ID
		}
	}

	maleCount := int64(0)
	femaleCount := int64(0)

	if maleID != 0 {
		c, e := os.profileService.CountBasicsCompletedByGender(ctx, maleID)
		if e != nil {
			return domain.PreregistrationStats{}, commonlogger.LogError(os.logger, "count male basics-completed", e, zap.Int16("genderID", maleID))
		}

		maleCount = c
	}

	if femaleID != 0 {
		c, e := os.profileService.CountBasicsCompletedByGender(ctx, femaleID)
		if e != nil {
			return domain.PreregistrationStats{}, commonlogger.LogError(os.logger, "count female basics-completed", e, zap.Int16("genderID", femaleID))
		}

		femaleCount = c
	}

	total, err := os.profileService.CountBasicsCompleted(ctx)
	if err != nil {
		return domain.PreregistrationStats{}, commonlogger.LogError(os.logger, "count total basics-completed", err)
	}

	return domain.PreregistrationStats{
		MaleCount:    maleCount,
		FemaleCount:  femaleCount,
		MaxTotal:     os.maxTotalParticipants,
		MaxMale:      os.maxMaleParticipants,
		MaxFemale:    os.maxFemaleParticipants,
		CapEnforced:  os.enablePreregCap,
		TotalCurrent: total,
	}, nil
}
