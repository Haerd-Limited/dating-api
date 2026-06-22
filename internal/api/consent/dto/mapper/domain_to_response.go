package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/consent/dto"
	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
)

func DomainsToResponse(consents []domain.Consent) dto.ListConsentsResponse {
	return DomainsToListResponse(consents)
}
