package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func MapMatchEntitiesToDomain(matchEntities []*entity.Match) []domain.Match {
	var matches []domain.Match
	for _, matchEntity := range matchEntities {
		matches = append(matches, MapMatchEntityToDomain(matchEntity))
	}

	return matches
}

func MapMatchEntityToDomain(matchEntity *entity.Match) domain.Match {
	var revealedAt time.Time
	if matchEntity.RevealedAt.Valid {
		revealedAt = matchEntity.RevealedAt.Time
	}

	return domain.Match{
		ID:         matchEntity.ID,
		UserA:      matchEntity.UserA,
		UserB:      matchEntity.UserB,
		CreatedAt:  matchEntity.CreatedAt,
		RevealedAt: revealedAt,
	}
}
