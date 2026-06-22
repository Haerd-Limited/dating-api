package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/consent/dto"
	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
)

func RecordRequestToDomain(req dto.RecordConsentRequest, userID string, ip, userAgent *string) domain.RecordRequest {
	return domain.RecordRequest{
		UserID:    userID,
		Type:      req.Type,
		Version:   req.Version,
		Accepted:  req.Accepted,
		IP:        ip,
		UserAgent: userAgent,
	}
}

func DomainToResponse(consent domain.Consent) dto.ConsentResponse {
	resp := dto.ConsentResponse{
		Type:       consent.Type,
		Version:    consent.Version,
		Accepted:   consent.Accepted,
		AcceptedAt: consent.AcceptedAt.UTC().Format(time.RFC3339),
	}

	if consent.RevokedAt != nil {
		revoked := consent.RevokedAt.UTC().Format(time.RFC3339)
		resp.RevokedAt = &revoked
	}

	return resp
}

func DomainsToListResponse(consents []domain.Consent) dto.ListConsentsResponse {
	out := make([]dto.ConsentResponse, 0, len(consents))
	for _, c := range consents {
		out = append(out, DomainToResponse(c))
	}

	return dto.ListConsentsResponse{Consents: out}
}
