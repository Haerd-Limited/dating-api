//go:generate mockgen -source=service.go -destination=service_mock.go -package=auditlog
package auditlog

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
	auditstorage "github.com/Haerd-Limited/dating-api/internal/auditlog/storage"
)

type Service interface {
	Record(ctx context.Context, entry domain.Entry) error
}

type service struct {
	logger *zap.Logger
	repo   auditstorage.Repository
}

func NewService(logger *zap.Logger, repo auditstorage.Repository) Service {
	return &service{
		logger: logger,
		repo:   repo,
	}
}

func (s *service) Record(ctx context.Context, entry domain.Entry) error {
	if entry.OccurredAt.IsZero() {
		entry.OccurredAt = time.Now().UTC()
	}

	err := s.repo.Insert(ctx, entry)
	if err != nil {
		s.logger.Warn("failed to record admin audit log",
			zap.String("method", entry.Method),
			zap.String("path", entry.Path),
			zap.Error(err))

		return err
	}

	return nil
}
