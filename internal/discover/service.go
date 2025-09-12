package discover

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/profilecard"
)

type Service interface {
	GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) ([]profilecard.ProfileCard, error)
}

type service struct {
	logger         *zap.Logger
	profileService profile.Service
	discoverRepo   storage.DiscoverRepository
}

func NewDiscoverService(
	logger *zap.Logger,
	profileService profile.Service,
	discoverRepo storage.DiscoverRepository,
) Service {
	return &service{
		logger:         logger,
		profileService: profileService,
		discoverRepo:   discoverRepo,
	}
}

// todo: add filters like  age, race, distance, age
func (s *service) GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) ([]profilecard.ProfileCard, error) {
	candidates, err := s.discoverRepo.GetCandidates(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate IDs userID=%s limit=%v offset=%v: %w", userID, limit, offset, err)
	}

	var profiles []profilecard.ProfileCard

	for _, candidate := range candidates {
		var profileErr error

		p, profileErr := s.profileService.GetProfileCard(ctx, candidate.UserID)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to get profile card userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}
