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
	GetRetentionStats(ctx context.Context, from, to time.Time) (domain.RetentionStats, error)
	GetRetentionCohorts(ctx context.Context, signupDate time.Time, daysAfter int) (domain.RetentionCohort, error)
	GetUserRetentionProfile(ctx context.Context, userID string) (domain.UserRetentionProfile, error)
	GetGlobalRetentionStats(ctx context.Context, date time.Time) (domain.RetentionStats, error)
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

func (s *service) GetRetentionStats(ctx context.Context, from, to time.Time) (domain.RetentionStats, error) {
	// Get DAU, WAU, MAU for the date range
	date := from

	dau, err := s.repo.GetDailyActiveUsers(ctx, date)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get daily active users: %w", err)
	}

	weekStart := startOfWeek(date)

	wau, err := s.repo.GetWeeklyActiveUsers(ctx, weekStart)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get weekly active users: %w", err)
	}

	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

	mau, err := s.repo.GetMonthlyActiveUsers(ctx, monthStart)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get monthly active users: %w", err)
	}

	totalUsers, err := s.repo.GetTotalUsers(ctx)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get total users: %w", err)
	}

	avgTimeToFirstReturn, err := s.repo.GetAverageTimeToFirstReturn(ctx, from, to)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get average time to first return: %w", err)
	}

	return domain.RetentionStats{
		TotalUsers:               totalUsers,
		DAU:                      dau,
		WAU:                      wau,
		MAU:                      mau,
		AverageTimeToFirstReturn: avgTimeToFirstReturn,
	}, nil
}

func (s *service) GetRetentionCohorts(ctx context.Context, signupDate time.Time, daysAfter int) (domain.RetentionCohort, error) {
	cohortSize, _, err := s.repo.GetRetentionCohort(ctx, signupDate, daysAfter)
	if err != nil {
		return domain.RetentionCohort{}, fmt.Errorf("get retention cohort: %w", err)
	}

	var day1Retention, day7Retention, day30Retention float64

	// Calculate Day 1 retention
	_, day1Returned, err := s.repo.GetRetentionCohort(ctx, signupDate, 1)
	if err == nil && cohortSize > 0 {
		day1Retention = float64(day1Returned) / float64(cohortSize) * 100
	}

	// Calculate Day 7 retention
	_, day7Returned, err := s.repo.GetRetentionCohort(ctx, signupDate, 7)
	if err == nil && cohortSize > 0 {
		day7Retention = float64(day7Returned) / float64(cohortSize) * 100
	}

	// Calculate Day 30 retention
	_, day30Returned, err := s.repo.GetRetentionCohort(ctx, signupDate, 30)
	if err == nil && cohortSize > 0 {
		day30Retention = float64(day30Returned) / float64(cohortSize) * 100
	}

	return domain.RetentionCohort{
		SignupDate:     signupDate,
		Day1Retention:  day1Retention,
		Day7Retention:  day7Retention,
		Day30Retention: day30Retention,
		CohortSize:     cohortSize,
	}, nil
}

func (s *service) GetUserRetentionProfile(ctx context.Context, userID string) (domain.UserRetentionProfile, error) {
	signupDate, err := s.repo.GetUserSignupDate(ctx, userID)
	if err != nil {
		return domain.UserRetentionProfile{}, fmt.Errorf("get user signup date: %w", err)
	}

	firstAppOpen, err := s.repo.GetFirstAppOpen(ctx, userID)

	var firstOpenTime *time.Time

	if err == nil && firstAppOpen != nil {
		firstOpenTime = &firstAppOpen.OccurredAt
	}

	latestAppOpen, err := s.repo.GetLatestAppOpen(ctx, userID)

	var latestOpenTime *time.Time

	if err == nil && latestAppOpen != nil {
		latestOpenTime = &latestAppOpen.OccurredAt
	}

	// Get total app opens (all time)
	totalAppOpens, err := s.repo.GetSessionFrequency(ctx, userID, time.Time{}, time.Now().UTC())
	if err != nil {
		totalAppOpens = 0 // Default to 0 if error
	}

	avgTimeBetweenOpens, err := s.repo.GetAverageTimeBetweenOpens(ctx, userID)
	if err != nil {
		// Not an error if user has < 2 opens
		avgTimeBetweenOpens = nil
	}

	var daysSinceLastOpen *int64

	if latestOpenTime != nil {
		days := int64(time.Since(*latestOpenTime).Hours() / 24)
		daysSinceLastOpen = &days
	}

	return domain.UserRetentionProfile{
		UserID:                  userID,
		SignupDate:              signupDate,
		FirstAppOpen:            firstOpenTime,
		LatestAppOpen:           latestOpenTime,
		TotalAppOpens:           totalAppOpens,
		AverageTimeBetweenOpens: avgTimeBetweenOpens,
		DaysSinceLastOpen:       daysSinceLastOpen,
	}, nil
}

func (s *service) GetGlobalRetentionStats(ctx context.Context, date time.Time) (domain.RetentionStats, error) {
	dau, err := s.repo.GetDailyActiveUsers(ctx, date)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get daily active users: %w", err)
	}

	weekStart := startOfWeek(date)

	wau, err := s.repo.GetWeeklyActiveUsers(ctx, weekStart)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get weekly active users: %w", err)
	}

	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

	mau, err := s.repo.GetMonthlyActiveUsers(ctx, monthStart)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get monthly active users: %w", err)
	}

	totalUsers, err := s.repo.GetTotalUsers(ctx)
	if err != nil {
		return domain.RetentionStats{}, fmt.Errorf("get total users: %w", err)
	}

	return domain.RetentionStats{
		TotalUsers: totalUsers,
		DAU:        dau,
		WAU:        wau,
		MAU:        mau,
	}, nil
}

func startOfWeek(t time.Time) time.Time {
	// Use Monday=0
	offset := (int(t.Weekday()) + 6) % 7
	day := t.AddDate(0, 0, -offset)

	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}
