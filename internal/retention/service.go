package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type Service interface {
	PurgeOnce(ctx context.Context) (Stats, error)
}

type Stats struct {
	VerificationCodesDeleted    int64 `json:"verification_codes_deleted"`
	EventsDeleted               int64 `json:"events_deleted"`
	VerificationAttemptsDeleted int64 `json:"verification_attempts_deleted"`
	RefreshTokensDeleted        int64 `json:"refresh_tokens_deleted"`
}

type service struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewService(logger *zap.Logger, db *sqlx.DB) Service {
	return &service{
		db:     db,
		logger: logger,
	}
}

func (s *service) PurgeOnce(ctx context.Context) (Stats, error) {
	var stats Stats

	now := time.Now().UTC()

	vcDeleted, err := s.purgeBatched(ctx, `
		DELETE FROM verification_codes
		 WHERE ctid IN (
			SELECT ctid FROM verification_codes
			 WHERE created_at < $1
			 LIMIT $2
		 )`, now.Add(-constants.RetentionVerificationCodes))
	if err != nil {
		s.logger.Warn("retention purge verification_codes failed", zap.Error(err))
	} else {
		stats.VerificationCodesDeleted = vcDeleted
	}

	eventsDeleted, err := s.purgeBatched(ctx, `
		DELETE FROM events
		 WHERE ctid IN (
			SELECT ctid FROM events
			 WHERE occurred_at < $1
			 LIMIT $2
		 )`, now.Add(-constants.RetentionEvents))
	if err != nil {
		s.logger.Warn("retention purge events failed", zap.Error(err))
	} else {
		stats.EventsDeleted = eventsDeleted
	}

	vaDeleted, err := s.purgeBatched(ctx, `
		DELETE FROM verification_attempts
		 WHERE ctid IN (
			SELECT ctid FROM verification_attempts
			 WHERE created_at < $1
			   AND status IN ('failed', 'needs_review')
			 LIMIT $2
		 )`, now.Add(-constants.RetentionVerificationAttempts))
	if err != nil {
		s.logger.Warn("retention purge verification_attempts failed", zap.Error(err))
	} else {
		stats.VerificationAttemptsDeleted = vaDeleted
	}

	rtDeleted, err := s.purgeBatched(ctx, `
		DELETE FROM refresh_tokens
		 WHERE ctid IN (
			SELECT ctid FROM refresh_tokens
			 WHERE expires_at < $1
			 LIMIT $2
		 )`, now.Add(-constants.RetentionExpiredRefreshTokensGrace))
	if err != nil {
		s.logger.Warn("retention purge refresh_tokens failed", zap.Error(err))
	} else {
		stats.RefreshTokensDeleted = rtDeleted
	}

	return stats, nil
}

func (s *service) purgeBatched(ctx context.Context, query string, cutoff time.Time) (int64, error) {
	var total int64

	for {
		res, err := s.db.ExecContext(ctx, query, cutoff, constants.RetentionPurgeBatchSize)
		if err != nil {
			return total, fmt.Errorf("batched delete: %w", err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return total, fmt.Errorf("rows affected: %w", err)
		}

		total += n

		if n == 0 {
			break
		}
	}

	return total, nil
}
