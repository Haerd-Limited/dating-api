package safety

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
)

func TestParseSuspendUntil(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	past := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	_, err := parseSuspendUntil(nil)
	assert.ErrorIs(t, err, ErrInvalidSuspendUntil)

	_, err = parseSuspendUntil(map[string]any{"suspend_until": past})
	assert.ErrorIs(t, err, ErrInvalidSuspendUntil)

	parsed, err := parseSuspendUntil(map[string]any{"suspend_until": future})
	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.True(t, parsed.After(time.Now().UTC()))
}

func TestExtractWarningMessage(t *testing.T) {
	t.Parallel()

	message, err := extractWarningMessage(safetydomain.ResolveReportRequest{
		ActionData: map[string]any{"warning_message": "Please follow community guidelines"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Please follow community guidelines", message)

	notes := "fallback note"
	message, err = extractWarningMessage(safetydomain.ResolveReportRequest{Notes: &notes})
	require.NoError(t, err)
	assert.Equal(t, "fallback note", message)

	_, err = extractWarningMessage(safetydomain.ResolveReportRequest{})
	assert.ErrorIs(t, err, ErrMissingWarningMessage)
}
