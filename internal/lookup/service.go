package lookup

import (
	"context"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/lookup/domain"
	"github.com/Haerd-Limited/dating-api/internal/lookup/mapper"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
)

type Service interface {
	GetPrompts(ctx context.Context) ([]domain.Prompt, error)
}

type lookupService struct {
	logger     *zap.Logger
	lookupRepo lookupstorage.LookupRepository
}

func NewLookupService(
	logger *zap.Logger,
	lookupRepo lookupstorage.LookupRepository,
) Service {
	return &lookupService{
		logger:     logger,
		lookupRepo: lookupRepo,
	}
}

func (s *lookupService) GetPrompts(ctx context.Context) ([]domain.Prompt, error) {
	prompts, err := s.lookupRepo.GetPrompts(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPromptsToDomain(prompts), nil
}
