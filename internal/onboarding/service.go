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
	Register(ctx context.Context, register domain.Register) (domain.RegisterResult, error)
	Basics(ctx context.Context, basics domain.Basics) (domain.BasicsResult, error)
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

func (os *onboardingService) Register(ctx context.Context, register domain.Register) (domain.RegisterResult, error) {
	// Call user service and insert into user service
	userID, err := os.userService.CreateUser(ctx, userdomain.User{
		Email:          register.Email,
		PhoneNumber:    register.PhoneNumber,
		FirstName:      register.FirstName,
		LastName:       register.LastName,
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

func (os *onboardingService) Basics(ctx context.Context, basics domain.Basics) (domain.BasicsResult, error) {
	userProfileEntity, err := os.repo.GetUserProfileByUserID(ctx, basics.UserID)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to get user profile by userID: %w", err)
	}

	userProfile := mapper.MapUserProfileToDomain(userProfileEntity)

	// todo: add checks to see if these ID's even exists
	userProfile.GenderID = basics.GenderID
	userProfile.DatingIntentionID = basics.DatingIntentionID
	userProfile.HeightCM = basics.HeightCm
	userProfile.Birthdate = basics.Birthdate

	UpdatedUserProfileEntity, err := mapper.MapProfileToEntity(userProfile.UserID, userProfile)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = os.repo.UpdateUserProfile(ctx, UpdatedUserProfileEntity)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to update user profile: %w", err)
	}

	userDomain, err := os.userService.GetUser(ctx, basics.UserID)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to get user details: %w", err)
	}

	userDomain.OnboardingStep = string(domain.OnboardingStepsLocation)

	err = os.userService.UpdateUser(ctx, userDomain)
	if err != nil {
		return domain.BasicsResult{}, fmt.Errorf("failed to update user onboarding step: %w", err)
	}

	return domain.BasicsResult{
		OnboardingSteps: domain.OnboardingStepsLocation.GenerateOnboardingSteps(),
	}, nil
}
