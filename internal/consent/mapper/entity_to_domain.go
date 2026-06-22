package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func EntitiesToDomain(entities []*entity.UserConsent) []domain.Consent {
	out := make([]domain.Consent, 0, len(entities))
	for _, e := range entities {
		out = append(out, EntityToDomain(e))
	}

	return out
}
