package domain

import "time"

type Period struct {
	Start time.Time
	End   time.Time
}

type GlobalWeekly struct {
	WeekStart            time.Time
	MessagesSent         int64
	VoiceMessagesSent    int64
	TopPromptID          *int64
	TopPromptLikeRate    *float64
	LikeToMatchRatePct   *float64 // placeholder, computed if available
	BusiestHourUTC       *int
	VerifiedVsUnverified *float64 // placeholder
}

type UserWeekly struct {
	WeekStart          time.Time
	LikesSent          int64
	LikesReceived      int64
	MatchesCreated     int64
	MessagesSent       int64
	VoiceMessagesSent  int64
	TopPromptID        *int64
	AvgResponseTimeSec *float64 // placeholder
}

type Wrapped struct {
	Year                 int
	TotalSwipes          int64
	TotalLikesSent       int64
	TotalMatches         int64
	TotalMessages        int64
	TopPromptID          *int64
	BestMonth            *int
	ConversationDepthPct *float64 // placeholder
	MatchToConvoRatePct  *float64 // placeholder
}

type RetentionStats struct {
	TotalUsers               int64
	DAU                      int64 // Daily Active Users
	WAU                      int64 // Weekly Active Users
	MAU                      int64 // Monthly Active Users
	AverageTimeToFirstReturn *time.Duration
}

type RetentionCohort struct {
	SignupDate     time.Time
	Day1Retention  float64 // % who returned on day 1
	Day7Retention  float64 // % who returned within 7 days
	Day30Retention float64 // % who returned within 30 days
	CohortSize     int64
}

type UserRetentionProfile struct {
	UserID                  string
	SignupDate              time.Time
	FirstAppOpen            *time.Time
	LatestAppOpen           *time.Time
	TotalAppOpens           int64
	AverageTimeBetweenOpens *time.Duration
	DaysSinceLastOpen       *int64
}
