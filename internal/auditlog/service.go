//go:generate mockgen -source=service.go -destination=service_mock.go -package=auditlog
package auditlog

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
	auditstorage "github.com/Haerd-Limited/dating-api/internal/auditlog/storage"
)

type Service interface {
	Record(ctx context.Context, entry domain.Entry) error
	ListEvents(ctx context.Context, filter domain.ListFilter) ([]domain.EventRow, error)
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

func (s *service) ListEvents(ctx context.Context, filter domain.ListFilter) ([]domain.EventRow, error) {
	entries, err := s.repo.ListEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	out := make([]domain.EventRow, 0, len(entries))

	for _, e := range entries {
		if !domain.IsMeaningfulAction(e.Method, e.Path) {
			continue
		}

		out = append(out, domain.EntryToEventRow(e))
	}

	return out, nil
}
