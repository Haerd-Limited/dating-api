package interaction

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/internal/interaction/mapper"
	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/profilecard"
)

type Service interface {
	CreateSwipe(ctx context.Context, swipe domain.Swipe) error
	GetLikes(ctx context.Context, userID, direction string, offset, limit int) ([]profilecard.ProfileCard, error)
	GetMatches(ctx context.Context, userID string) ([]domain.Match, error)
}

type service struct {
	logger          *zap.Logger
	profileService  profile.Service
	interactionRepo storage.InteractionRepository
}

func NewInteractionService(
	logger *zap.Logger,
	interactionRepo storage.InteractionRepository,
	profileService profile.Service,
) Service {
	return &service{
		logger:          logger,
		interactionRepo: interactionRepo,
		profileService:  profileService,
	}
}

var ErrInvalidDirection = fmt.Errorf("invalid direction")

func (is *service) CreateSwipe(ctx context.Context, swipe domain.Swipe) error {
	//todo:if like or super like, notifiy
	err := is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe))
	if err != nil {
		return fmt.Errorf("failed to insert swipe userID=%s : %w", swipe.UserID, err)
	}

	matchable, err := is.interactionRepo.CheckIfMatchable(ctx, swipe.UserID, swipe.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to check if matchable userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
	}

	if !matchable {
		return nil
	}

	// Create a match (normalize order to keep uniqueness deterministic)
	a, b := swipe.UserID, swipe.TargetUserID
	if b < a {
		a, b = b, a
	}

	err = is.interactionRepo.CreateMatch(ctx, entity.Match{
		UserA: a,
		UserB: b,
	})
	if err != nil {
		return fmt.Errorf("failed to created match userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
	}
	//todo:send notification to targetuser

	return nil
}

func (is *service) GetLikes(ctx context.Context, userID, direction string, offset, limit int) ([]profilecard.ProfileCard, error) {
	var likesUserIDs []string

	var profiles []profilecard.ProfileCard

	var err error

	switch direction {
	case "incoming":
		likesUserIDs, err = is.interactionRepo.GetIncomingLikes(ctx, userID, limit, offset)
	default:
		return nil, ErrInvalidDirection
	}

	if err != nil {
		return nil, err
	}

	for _, id := range likesUserIDs {
		p, profileErr := is.profileService.GetProfileCard(ctx, id)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to get profile card userID=%s profileUserID=%s: %w", userID, id, profileErr)
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}

func (is *service) GetMatches(ctx context.Context, userID string) ([]domain.Match, error) {
	matchEntities, err := is.interactionRepo.GetMatches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches userID=%s: %w", userID, err)
	}

	if len(matchEntities) == 0 {
		return nil, nil
	}

	var matches []domain.Match

	for _, matchEntity := range matchEntities {
		var matchedUserID string
		if matchEntity.UserA == userID {
			matchedUserID = matchEntity.UserB
		} else {
			matchedUserID = matchEntity.UserA
		}
		// get display name
		profileCard, profileErr := is.profileService.GetProfileCard(ctx, matchedUserID)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to get profile card userID=%s profileUserID=%s: %w", userID, matchedUserID, profileErr)
		}
		// todo: Get latest message
		// todo:  calculate reveal progress

		// set reveal
		matches = append(matches, domain.Match{
			UserID:         matchedUserID,
			DisplayName:    profileCard.DisplayName,
			Emoji:          "😄", // todo: update register or edit profile to allow user to set emoji. this emoji is default for now
			MessagePreview: "To be implemented",
			Reveal:         false,
			RevealProgress: 0,
		})
	}

	return matches, nil
}
