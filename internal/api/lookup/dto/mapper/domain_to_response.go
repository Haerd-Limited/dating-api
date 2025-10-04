package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/lookup/dto"
	"github.com/Haerd-Limited/dating-api/internal/lookup/domain"
)

func MapToGetPromptsResponse(domainPrompts []domain.Prompt) dto.GetPromptsResponse {
	var prompts []dto.Prompt
	for _, prompt := range domainPrompts {
		prompts = append(prompts, dto.Prompt{
			ID:       prompt.ID,
			Label:    prompt.Label,
			Key:      prompt.Key,
			Category: prompt.Category,
		})
	}

	return dto.GetPromptsResponse{
		Prompts: prompts,
	}
}
