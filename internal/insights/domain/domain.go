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
