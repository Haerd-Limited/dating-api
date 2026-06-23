package storage

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

// Compile-time assertion that the concrete repository satisfies the interface
// (incl. the new GetAccountGateState / UpdateAccountStatus methods).
var _ UserRepository = (*userRepository)(nil)

// TestAccountGateStateRoundTripIntegration exercises the single-query gate read
// and the tx-aware status write against a migrated database. It mirrors the
// skip-by-default convention used by the auth storage integration test.
func TestAccountGateStateRoundTripIntegration(t *testing.T) {
	t.Skip("integration test: requires migrated database")

	db, err := sqlx.Connect("postgres", "")
	require.NoError(t, err)

	repo := NewUserRepository(db).(*userRepository)

	const userID = "00000000-0000-0000-0000-000000000001"

	// Default state: active, no suspension, no pending warning.
	state, err := repo.GetAccountGateState(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatusActive, state.Status)
	assert.Nil(t, state.SuspendedUntil)
	assert.False(t, state.HasPendingWarning)

	// Suspend the user without an outer transaction (tx == nil path).
	until := time.Now().UTC().Add(24 * time.Hour)
	reason := "spam"
	require.NoError(t, repo.UpdateAccountStatus(
		context.Background(),
		userID,
		domain.AccountStatusSuspended,
		&until,
		&reason,
		nil,
	))

	state, err = repo.GetAccountGateState(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatusSuspended, state.Status)
	require.NotNil(t, state.SuspendedUntil)
	assert.WithinDuration(t, until, *state.SuspendedUntil, time.Second)
	require.NotNil(t, state.Reason)
	assert.Equal(t, reason, *state.Reason)
}

// TestUpdateAccountStatusMissingUserIntegration documents that updating an
// unknown user is a no-op (no row affected, no error), per the raw UPDATE.
func TestUpdateAccountStatusMissingUserIntegration(t *testing.T) {
	t.Skip("integration test: requires migrated database")

	db, err := sqlx.Connect("postgres", "")
	require.NoError(t, err)

	repo := NewUserRepository(db).(*userRepository)

	err = repo.UpdateAccountStatus(
		context.Background(),
		"00000000-0000-0000-0000-0000000000ff",
		domain.AccountStatusBanned,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
}
