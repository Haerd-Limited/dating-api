package discover

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"github.com/Haerd-Limited/dating-api/internal/matching"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

type Service interface {
	GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) ([]profilecard.ProfileCard, error)
	GetVoiceWorthHearing(ctx context.Context, userID string) ([]profilecard.ProfileCard, error)
	GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
}

type service struct {
	logger          *zap.Logger
	profileService  profile.Service
	matchingService matching.Service
	discoverRepo    storage.DiscoverRepository
}

func NewDiscoverService(
	logger *zap.Logger,
	profileService profile.Service,
	matchingService matching.Service,
	discoverRepo storage.DiscoverRepository,
) Service {
	return &service{
		logger:          logger,
		profileService:  profileService,
		matchingService: matchingService,
		discoverRepo:    discoverRepo,
	}
}

const minOverlap = 5

func (s *service) AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error) {
	return s.discoverRepo.AlreadyInteracted(ctx, userID, targetUserID)
}

// todo(high-priority): add filters like  age range, dating intentions, distance, age
func (s *service) GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) ([]profilecard.ProfileCard, error) {
	candidates, err := s.discoverRepo.GetDiscoverFeedCandidates(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate IDs userID=%s limit=%v offset=%v: %w", userID, limit, offset, err)
	}

	// Get current user's location for distance calculation
	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user profile userID=%s: %w", userID, err)
	}

	var profiles []profilecard.ProfileCard

	for _, candidate := range candidates {
		var profileErr error

		p, profileErr := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to get profile card userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		p.MatchSummary, profileErr = s.computeMatch(ctx, userID, candidate.UserID, minOverlap)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to compute match userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}

func (s *service) GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error) {
	numberOfOppositeGenderProfiles, err := s.discoverRepo.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if numberOfOppositeGenderProfiles < constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, nil
	}

	profiles, err := s.GetVoiceWorthHearing(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get voice worth hearing userID=%s: %w", userID, err)
	}

	ids := make([]string, 0, len(profiles))
	for _, p := range profiles {
		ids = append(ids, p.UserID)
	}

	return ids, nil
}

// todo(high-priority): update to refresh weekly. maybe make a table to store user's voices to be heard. then have cron job recalculate and update weekly
// todo (high-priority): update to be more tailored to user's preferences e.g. race age etc
func (s *service) GetVoiceWorthHearing(ctx context.Context, userID string) ([]profilecard.ProfileCard, error) {
	candidates, err := s.discoverRepo.GetVoiceWorthHearing(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get candidates userID=%s: %w", userID, err)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Get current user's location for distance calculation
	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user profile userID=%s: %w", userID, err)
	}

	var profiles []profilecard.ProfileCard

	for _, candidate := range candidates {
		var profileErr error

		alreadyInteracted, profileErr := s.discoverRepo.AlreadyInteracted(ctx, userID, candidate.UserID)
		if profileErr != nil {
			return nil, fmt.Errorf("already interacted userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		if alreadyInteracted {
			continue
		}

		p, profileErr := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
		if profileErr != nil {
			return nil, fmt.Errorf("get profile card userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		var likeCount int64

		likeCount, profileErr = s.discoverRepo.GetLikeAndSuperlikeCount(ctx, candidate.UserID)
		if profileErr != nil {
			return nil, fmt.Errorf("get like and superlike count userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		p.LikeCount = &likeCount

		p.MatchSummary, profileErr = s.computeMatch(ctx, userID, candidate.UserID, minOverlap)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to compute match userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}

func (s *service) computeMatch(ctx context.Context, userID string, candidateID string, minOverlap int) (*profilecard.MatchSummary, error) {
	matchSummary, err := s.matchingService.ComputeMatch(ctx, userID, candidateID, minOverlap)
	if err != nil {
		return nil, err
	}

	result := &profilecard.MatchSummary{
		MatchPercent: matchSummary.MatchPercent,
		OverlapCount: matchSummary.OverlapCount,
		Badges:       nil,
		HiddenReason: matchSummary.HiddenReason,
	}
	for _, badge := range matchSummary.Badges {
		result.Badges = append(result.Badges, profilecard.MatchBadge{
			QuestionID:    badge.QuestionID,
			QuestionText:  badge.QuestionText,
			PartnerAnswer: badge.PartnerAnswer,
			Weight:        badge.Weight,
		})
	}

	return result, nil
}
