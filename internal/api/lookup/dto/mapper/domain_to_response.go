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

func MapToGetLanguagesResponse(domainLangs []domain.Language) dto.GetLanguagesResponse {
	var languages []dto.Language
	for _, lang := range domainLangs {
		languages = append(languages, dto.Language{
			ID:    lang.ID,
			Label: lang.Label,
		})
	}

	return dto.GetLanguagesResponse{
		Languages: languages,
	}
}

func MapToGetReligionsResponse(domainReligions []domain.Religion) dto.GetReligionsResponse {
	var religions []dto.Religion
	for _, religion := range domainReligions {
		religions = append(religions, dto.Religion{
			ID:    religion.ID,
			Label: religion.Label,
		})
	}

	return dto.GetReligionsResponse{
		Religions: religions,
	}
}

func MapToGetPoliticalBeliefsResponse(domainPoliticalBeliefs []domain.PoliticalBelief) dto.GetPoliticalBeliefsResponse {
	var politicalBeliefs []dto.PoliticalBelief
	for _, politicalBelief := range domainPoliticalBeliefs {
		politicalBeliefs = append(politicalBeliefs, dto.PoliticalBelief{
			ID:    politicalBelief.ID,
			Label: politicalBelief.Label,
		})
	}

	return dto.GetPoliticalBeliefsResponse{
		PoliticalBeliefs: politicalBeliefs,
	}
}
