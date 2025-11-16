package insights

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/insights/domain"
	insstorage "github.com/Haerd-Limited/dating-api/internal/insights/storage"
)

type Service interface {
	GetPublicWeekly(ctx context.Context, weekStart time.Time) (domain.GlobalWeekly, error)
	GetUserWeekly(ctx context.Context, userID string, weekStart time.Time) (domain.UserWeekly, error)
	GetWrapped(ctx context.Context, userID string, year int) (domain.Wrapped, error)
}

type service struct {
	logger *zap.Logger
	repo   insstorage.Repository
}

func NewService(logger *zap.Logger, repo insstorage.Repository) Service {
	return &service{
		logger: logger,
		repo:   repo,
	}
}

func weekRange(weekStart time.Time) (time.Time, time.Time) {
	ws := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	return ws, ws.AddDate(0, 0, 7)
}

func (s *service) GetPublicWeekly(ctx context.Context, weekStart time.Time) (domain.GlobalWeekly, error) {
	from, to := weekRange(weekStart)

	msgs, voice, err := s.repo.CountMessages(ctx, from, to)
	if err != nil {
		return domain.GlobalWeekly{}, fmt.Errorf("count messages: %w", err)
	}

	topPromptID, topRate, err := s.repo.TopPromptWeekly(ctx, from)
	if err != nil {
		// log only; not fatal
		s.logger.Sugar().Warnw("top prompt weekly failed", "error", err)
	}

	return domain.GlobalWeekly{
		WeekStart:         from,
		MessagesSent:      msgs,
		VoiceMessagesSent: voice,
		TopPromptID:       topPromptID,
		TopPromptLikeRate: topRate,
	}, nil
}

func (s *service) GetUserWeekly(ctx context.Context, userID string, weekStart time.Time) (domain.UserWeekly, error) {
	from, to := weekRange(weekStart)

	likesSent, likesReceived, matches, messages, voice, err := s.repo.UserCounts(ctx, userID, from, to)
	if err != nil {
		return domain.UserWeekly{}, fmt.Errorf("user counts: %w", err)
	}

	return domain.UserWeekly{
		WeekStart:         from,
		LikesSent:         likesSent,
		LikesReceived:     likesReceived,
		MatchesCreated:    matches,
		MessagesSent:      messages,
		VoiceMessagesSent: voice,
	}, nil
}

func (s *service) GetWrapped(ctx context.Context, userID string, year int) (domain.Wrapped, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	likesSent, likesReceived, matches, messages, _, err := s.repo.UserCounts(ctx, userID, start, end)
	if err != nil {
		return domain.Wrapped{}, fmt.Errorf("user counts: %w", err)
	}

	// Rough best month by messages sent
	var bestMonth *int

	var bestVal int64

	for m := 1; m <= 12; m++ {
		mStart := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		mEnd := mStart.AddDate(0, 1, 0)

		_, _, _, monthlyMessages, _, e := s.repo.UserCounts(ctx, userID, mStart, mEnd)
		if e != nil {
			continue
		}

		if monthlyMessages > bestVal {
			bestVal = monthlyMessages
			mv := m
			bestMonth = &mv
		}
	}

	result := domain.Wrapped{
		Year:           year,
		TotalSwipes:    likesSent + likesReceived, // proxy without duplicates
		TotalLikesSent: likesSent,
		TotalMatches:   matches,
		TotalMessages:  messages,
		BestMonth:      bestMonth,
	}

	_ = s.repo.UpsertWrapped(ctx, userID, year, result)

	return result, nil
}
