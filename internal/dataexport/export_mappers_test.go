package dataexport

import (
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/dataexport/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func TestConsentsToExport(t *testing.T) {
	t.Parallel()

	revoked := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	acceptedAt := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)

	out := consentsToExport([]*entity.UserConsent{
		{
			ConsentType: "privacy_policy",
			Version:     "2026-05-28",
			Accepted:    true,
			AcceptedAt:  acceptedAt,
			IP:          null.StringFrom("203.0.113.1"),
			UserAgent:   null.StringFrom("Mozilla/5.0"),
		},
		{
			ConsentType: "terms_of_service",
			Version:     "2026-05-28",
			Accepted:    false,
			AcceptedAt:  acceptedAt,
			RevokedAt:   null.TimeFrom(revoked),
		},
	})

	require.Len(t, out, 2)
	assert.Equal(t, domain.ConsentExport{
		ConsentType: "privacy_policy",
		Version:     "2026-05-28",
		Accepted:    true,
		AcceptedAt:  acceptedAt,
	}, out[0])
	require.NotNil(t, out[1].RevokedAt)
	assert.Equal(t, revoked, *out[1].RevokedAt)
}

func TestDeviceTokensToExport(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 1, 8, 30, 0, 0, time.UTC)

	out := deviceTokensToExport([]*entity.DeviceToken{
		{
			Token:     "ExponentPushToken[abc123]",
			CreatedAt: null.TimeFrom(createdAt),
		},
	})

	require.Len(t, out, 1)
	assert.Equal(t, domain.DeviceTokenExport{
		Token:     "ExponentPushToken[abc123]",
		CreatedAt: createdAt,
	}, out[0])
}
