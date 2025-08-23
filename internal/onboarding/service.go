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
		OnboardingStep: string(register.OnboardingStep),
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
		OnboardingSteps: register.OnboardingStep.GenerateOnboardingSteps(),
		Content: domain.RegisterContent{
			DatingIntentions: mapper.MapDatingIntentionsToDomain(datingIntentionEntities),
			Genders:          mapper.MapGendersToDomain(genderEntities),
		},
	}, nil
}
