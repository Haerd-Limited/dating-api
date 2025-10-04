package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/lookup/domain"
)

func MapPromptsToDomain(g []*entity.PromptType) []domain.Prompt {
	if g == nil {
		return nil
	}

	var result []domain.Prompt

	for _, e := range g {
		result = append(result, domain.Prompt{
			ID:       e.ID,
			Label:    e.Label,
			Key:      e.Key,
			Category: e.Category,
		})
	}

	return result
}

func MapLanguagesToDomain(g []*entity.Language) []domain.Language {
	if g == nil {
		return nil
	}

	var result []domain.Language

	for _, e := range g {
		result = append(result, domain.Language{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}
