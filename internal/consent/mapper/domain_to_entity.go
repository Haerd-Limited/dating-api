package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func EntityToDomain(e *entity.UserConsent) domain.Consent {
	out := domain.Consent{
		UserID:     e.UserID,
		Type:       e.ConsentType,
		Version:    e.Version,
		Accepted:   e.Accepted,
		AcceptedAt: e.AcceptedAt,
	}

	if e.RevokedAt.Valid {
		t := e.RevokedAt.Time
		out.RevokedAt = &t
	}

	if e.IP.Valid {
		ip := e.IP.String
		out.IP = &ip
	}

	if e.UserAgent.Valid {
		ua := e.UserAgent.String
		out.UserAgent = &ua
	}

	return out
}

func RecordRequestToEntity(req domain.RecordRequest) *entity.UserConsent {
	out := &entity.UserConsent{
		UserID:      req.UserID,
		ConsentType: req.Type,
		Version:     req.Version,
		Accepted:    req.Accepted,
	}

	if req.IP != nil {
		out.IP = null.StringFrom(*req.IP)
	}

	if req.UserAgent != nil {
		out.UserAgent = null.StringFrom(*req.UserAgent)
	}

	return out
}
