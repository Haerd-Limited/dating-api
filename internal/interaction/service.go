package interaction

import (
	"context"

	"go.uber.org/zap"

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
	err := is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe))
	if err != nil {
		return err
	}

	return nil
}
