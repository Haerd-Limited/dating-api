package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	scoredomain "github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
)

func MapScoreDomainSnapShotToConversationDomain(ss scoredomain.ScoreSnapshot) *domain.ScoreSnapshot {
	return &domain.ScoreSnapshot{
		Threshold: ss.Threshold,
		Me:        ss.Me,
		Them:      ss.Them,
		Shared:    ss.Shared,
		CanReveal: ss.CanReveal,
		Revealed:  ss.Revealed,
	}
}
