package interaction

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/internal/interaction/mapper"
	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
)

type Service interface {
	CreateSwipe(ctx context.Context, swipe domain.Swipe) error
}

type service struct {
	logger          *zap.Logger
	interactionRepo storage.InteractionRepository
}

func NewInteractionService(logger *zap.Logger, interactionRepo storage.InteractionRepository) Service {
	return &service{
		logger:          logger,
		interactionRepo: interactionRepo,
	}
}

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
	//todo:send notification

	return nil
}
