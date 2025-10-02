package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

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

func MapEthnicityToDomain(g []*entity.Ethnicity) []domain.Ethnicity {
	if g == nil {
		return nil
	}

	var result []domain.Ethnicity

	for _, e := range g {
		result = append(result, domain.Ethnicity{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapEducationlevelsToDomain(g []*entity.EducationLevel) []domain.EducationLevel {
	if g == nil {
		return nil
	}

	var result []domain.EducationLevel

	for _, e := range g {
		result = append(result, domain.EducationLevel{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapPoliticalBeliefsToDomain(g []*entity.PoliticalBelief) []domain.PoliticalBelief {
	if g == nil {
		return nil
	}

	var result []domain.PoliticalBelief

	for _, e := range g {
		result = append(result, domain.PoliticalBelief{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapReligionsToDomain(g []*entity.Religion) []domain.Religion {
	if g == nil {
		return nil
	}

	var result []domain.Religion

	for _, e := range g {
		result = append(result, domain.Religion{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapHabitsToDomain(g []*entity.Habit) []domain.Habit {
	if g == nil {
		return nil
	}

	var result []domain.Habit

	for _, e := range g {
		result = append(result, domain.Habit{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

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

func MapGendersToDomain(g []*entity.Gender) []domain.Gender {
	if g == nil {
		return nil
	}

	var result []domain.Gender

	for _, e := range g {
		result = append(result, domain.Gender{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapDatingIntentionsToDomain(di []*entity.DatingIntention) []domain.DatingIntention {
	if di == nil {
		return nil
	}

	var result []domain.DatingIntention
	for _, e := range di {
		result = append(result, domain.DatingIntention{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}
