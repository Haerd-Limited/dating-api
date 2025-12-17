package dto

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/insights/domain"
)

type PublicWeeklyResponse struct {
	WeekStart         string   `json:"week_start"`
	MessagesSent      int64    `json:"messages_sent"`
	VoiceMessagesSent int64    `json:"voice_messages_sent"`
	TopPromptID       *int64   `json:"top_prompt_id,omitempty"`
	TopPromptLikeRate *float64 `json:"top_prompt_like_rate,omitempty"`
}

type UserWeeklyResponse struct {
	WeekStart         string `json:"week_start"`
	LikesSent         int64  `json:"likes_sent"`
	LikesReceived     int64  `json:"likes_received"`
	MatchesCreated    int64  `json:"matches_created"`
	MessagesSent      int64  `json:"messages_sent"`
	VoiceMessagesSent int64  `json:"voice_messages_sent"`
}

type WrappedResponse struct {
	Year           int    `json:"year"`
	TotalSwipes    int64  `json:"total_swipes"`
	TotalLikesSent int64  `json:"total_likes_sent"`
	TotalMatches   int64  `json:"total_matches"`
	TotalMessages  int64  `json:"total_messages"`
	BestMonth      *int   `json:"best_month,omitempty"`
	TopPromptID    *int64 `json:"top_prompt_id,omitempty"`
}

func MapGlobalWeekly(g domain.GlobalWeekly) PublicWeeklyResponse {
	return PublicWeeklyResponse{
		WeekStart:         g.WeekStart.Format("2006-01-02"),
		MessagesSent:      g.MessagesSent,
		VoiceMessagesSent: g.VoiceMessagesSent,
		TopPromptID:       g.TopPromptID,
		TopPromptLikeRate: g.TopPromptLikeRate,
	}
}

func MapUserWeekly(u domain.UserWeekly) UserWeeklyResponse {
	return UserWeeklyResponse{
		WeekStart:         u.WeekStart.Format("2006-01-02"),
		LikesSent:         u.LikesSent,
		LikesReceived:     u.LikesReceived,
		MatchesCreated:    u.MatchesCreated,
		MessagesSent:      u.MessagesSent,
		VoiceMessagesSent: u.VoiceMessagesSent,
	}
}

func MapWrapped(w domain.Wrapped) WrappedResponse {
	return WrappedResponse{
		Year:           w.Year,
		TotalSwipes:    w.TotalSwipes,
		TotalLikesSent: w.TotalLikesSent,
		TotalMatches:   w.TotalMatches,
		TotalMessages:  w.TotalMessages,
		BestMonth:      w.BestMonth,
		TopPromptID:    w.TopPromptID,
	}
}

type RetentionStatsResponse struct {
	TotalUsers               int64   `json:"total_users"`
	DAU                      int64   `json:"dau"`                                    // Daily Active Users
	WAU                      int64   `json:"wau"`                                    // Weekly Active Users
	MAU                      int64   `json:"mau"`                                    // Monthly Active Users
	AverageTimeToFirstReturn *string `json:"average_time_to_first_return,omitempty"` // ISO 8601 duration
}

type RetentionCohortResponse struct {
	SignupDate     string  `json:"signup_date"`      // ISO 8601 date
	Day1Retention  float64 `json:"day_1_retention"`  // Percentage
	Day7Retention  float64 `json:"day_7_retention"`  // Percentage
	Day30Retention float64 `json:"day_30_retention"` // Percentage
	CohortSize     int64   `json:"cohort_size"`
}

type UserRetentionProfileResponse struct {
	UserID                  string  `json:"user_id"`
	SignupDate              string  `json:"signup_date"`               // ISO 8601 datetime
	FirstAppOpen            *string `json:"first_app_open,omitempty"`  // ISO 8601 datetime
	LatestAppOpen           *string `json:"latest_app_open,omitempty"` // ISO 8601 datetime
	TotalAppOpens           int64   `json:"total_app_opens"`
	AverageTimeBetweenOpens *string `json:"average_time_between_opens,omitempty"` // ISO 8601 duration
	DaysSinceLastOpen       *int64  `json:"days_since_last_open,omitempty"`
}

func MapRetentionStats(r domain.RetentionStats) RetentionStatsResponse {
	var avgTimeStr *string

	if r.AverageTimeToFirstReturn != nil {
		hours := int64(r.AverageTimeToFirstReturn.Hours())
		minutes := int64(r.AverageTimeToFirstReturn.Minutes()) % 60
		seconds := int64(r.AverageTimeToFirstReturn.Seconds()) % 60
		s := formatDuration(hours, minutes, seconds)
		avgTimeStr = &s
	}

	return RetentionStatsResponse{
		TotalUsers:               r.TotalUsers,
		DAU:                      r.DAU,
		WAU:                      r.WAU,
		MAU:                      r.MAU,
		AverageTimeToFirstReturn: avgTimeStr,
	}
}

func MapRetentionCohort(c domain.RetentionCohort) RetentionCohortResponse {
	return RetentionCohortResponse{
		SignupDate:     c.SignupDate.Format("2006-01-02"),
		Day1Retention:  c.Day1Retention,
		Day7Retention:  c.Day7Retention,
		Day30Retention: c.Day30Retention,
		CohortSize:     c.CohortSize,
	}
}

func MapUserRetentionProfile(p domain.UserRetentionProfile) UserRetentionProfileResponse {
	var firstOpenStr, latestOpenStr *string

	if p.FirstAppOpen != nil {
		s := p.FirstAppOpen.Format(time.RFC3339)
		firstOpenStr = &s
	}

	if p.LatestAppOpen != nil {
		s := p.LatestAppOpen.Format(time.RFC3339)
		latestOpenStr = &s
	}

	var avgTimeStr *string

	if p.AverageTimeBetweenOpens != nil {
		hours := int64(p.AverageTimeBetweenOpens.Hours())
		minutes := int64(p.AverageTimeBetweenOpens.Minutes()) % 60
		seconds := int64(p.AverageTimeBetweenOpens.Seconds()) % 60
		s := formatDuration(hours, minutes, seconds)
		avgTimeStr = &s
	}

	return UserRetentionProfileResponse{
		UserID:                  p.UserID,
		SignupDate:              p.SignupDate.Format(time.RFC3339),
		FirstAppOpen:            firstOpenStr,
		LatestAppOpen:           latestOpenStr,
		TotalAppOpens:           p.TotalAppOpens,
		AverageTimeBetweenOpens: avgTimeStr,
		DaysSinceLastOpen:       p.DaysSinceLastOpen,
	}
}

func formatDuration(hours, minutes, seconds int64) string {
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	return d.String()
}
