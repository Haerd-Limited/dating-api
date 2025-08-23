package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

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
