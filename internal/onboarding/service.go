package onboarding

import (
	"go.uber.org/zap"

	onboardingstorage "github.com/Haerd-Limited/dating-api/internal/onboarding/storage"
)

type Service interface{}

type onboardingService struct {
	logger         *zap.Logger
	onboardingRepo onboardingstorage.OnboardingRepository
}

func NewOnboardingService(
	logger *zap.Logger,
	onboardingRepository onboardingstorage.OnboardingRepository,
) Service {
	return &onboardingService{
		logger:         logger,
		onboardingRepo: onboardingRepository,
	}
}
