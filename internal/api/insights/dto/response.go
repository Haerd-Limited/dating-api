package dto

import (
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
