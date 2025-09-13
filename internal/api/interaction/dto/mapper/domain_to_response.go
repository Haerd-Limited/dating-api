package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/interaction/dto"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
)

func MapMatchesToDto(matches []domain.Match) []dto.Match {
	if matches == nil {
		return []dto.Match{}
	}

	var matchesDto []dto.Match
	for _, match := range matches {
		matchesDto = append(matchesDto, MapMatchtoDto(match))
	}

	return matchesDto
}

func MapMatchtoDto(match domain.Match) dto.Match {
	return dto.Match{
		UserID:         match.UserID,
		DisplayName:    match.DisplayName,
		MessagePreview: match.MessagePreview,
		Emoji:          match.Emoji,
		Reveal:         match.Reveal,
		RevealProgress: match.RevealProgress,
	}
}
