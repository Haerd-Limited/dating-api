package dto

import (
	"testing"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/insights/domain"
)

func TestMapUserWeekly(t *testing.T) {
	ts := time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)
	in := domain.UserWeekly{
		WeekStart:         ts,
		LikesSent:         5,
		LikesReceived:     3,
		MatchesCreated:    2,
		MessagesSent:      10,
		VoiceMessagesSent: 4,
	}

	out := MapUserWeekly(in)
	if out.WeekStart != "2025-11-10" {
		t.Fatalf("expected week_start formatted, got %s", out.WeekStart)
	}

	if out.LikesSent != 5 || out.LikesReceived != 3 || out.MatchesCreated != 2 {
		t.Fatalf("unexpected aggregation mapping: %+v", out)
	}
}
