package onboarding

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
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
}

type onboardingService struct {
	logger      *zap.Logger
	repo        onboardingstorage.OnboardingRepository
	userService user.Service
	authService auth.Service
}

func NewOnboardingService(
	logger *zap.Logger,
	onboardingRepository onboardingstorage.OnboardingRepository,
	userService user.Service,
	authService auth.Service,
) Service {
	return &onboardingService{
		logger:      logger,
		repo:        onboardingRepository,
		userService: userService,
		authService: authService,
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
	genderEntities, err := os.repo.GetGenders(ctx)
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to get genders: %w", err)
	}

	datingIntentionEntities, err := os.repo.GetDatingIntentions(ctx)
	if err != nil {
		return domain.RegisterResult{}, fmt.Errorf("failed to get dating intentions: %w", err)
	}

	return domain.RegisterResult{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		OnboardingSteps: domain.OnboardingStepsBasics.GenerateOnboardingSteps(),
		Content: domain.RegisterContent{
			DatingIntentions: mapper.MapDatingIntentionsToDomain(datingIntentionEntities),
			Genders:          mapper.MapGendersToDomain(genderEntities),
		},
	}, nil
}

// TODO: FIX hitting the same endpoint multiple times advancing steps
func (os *onboardingService) Basics(ctx context.Context, basicDetails domain.Basics) (domain.BasicsResult, error) {
	userProfile, err := os.getUserProfile(ctx, basicDetails.UserID)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	// todo: add checks to see if these ID's even exists
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

	userProfile.City = &locationDetails.City
	userProfile.Country = &locationDetails.Country
	userProfile.Latitude = &locationDetails.Latitude
	userProfile.Longitude = &locationDetails.Longitude

	err = os.updateUserProfile(ctx, userProfile)
	if err != nil {
		return domain.LocationResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	onBoardingStep, err := os.bumpOnboardingStep(ctx, locationDetails.UserID)
	if err != nil {
		return domain.LocationResult{}, fmt.Errorf("failed to bump onboarding step: %w", err)
	}

	habits, err := os.GetHabits(ctx)
	if err != nil {
		return domain.LocationResult{}, err
	}

	return domain.LocationResult{
		OnboardingSteps: onBoardingStep.GenerateOnboardingSteps(),
		Content:         domain.LocationContent{Habits: habits},
	}, nil
}

func (os *onboardingService) GetHabits(ctx context.Context) ([]domain.Habit, error) {
	habitEntities, err := os.repo.GetHabits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits: %w", err)
	}

	return mapper.MapHabitsToDomain(habitEntities), nil
}

func (os *onboardingService) updateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error {
	updatedUserProfileEntity, err := mapper.MapProfileToEntity(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	os.logger.Info("updating user profile",
		zap.Any("geo", updatedUserProfileEntity.Geo),
		zap.Any("lat", userProfile.Latitude),
		zap.Any("long", userProfile.Longitude),
	)

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
