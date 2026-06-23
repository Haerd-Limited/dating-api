package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccountStateEffectiveStatus(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(24 * time.Hour)
	past := time.Now().UTC().Add(-24 * time.Hour)
	now := time.Now().UTC()

	assert.Equal(t, AccountStatusActive, (AccountState{Status: AccountStatusActive}).EffectiveStatus(now))
	assert.Equal(t, AccountStatusBanned, (AccountState{Status: AccountStatusBanned}).EffectiveStatus(now))

	activeAfterExpiry := AccountState{Status: AccountStatusSuspended, SuspendedUntil: &past}
	assert.Equal(t, AccountStatusActive, activeAfterExpiry.EffectiveStatus(now))

	stillSuspended := AccountState{Status: AccountStatusSuspended, SuspendedUntil: &future}
	assert.Equal(t, AccountStatusSuspended, stillSuspended.EffectiveStatus(now))
}
