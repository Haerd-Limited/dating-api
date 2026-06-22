package storage

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestDeleteVerificationCodesForUserIntegration(t *testing.T) {
	t.Skip("integration test: requires migrated database")

	db, err := sqlx.Connect("postgres", "")
	require.NoError(t, err)

	repo := NewAuthRepository(db)
	phone := "+441234567890"
	email := "user@example.com"
	err = repo.DeleteVerificationCodesForUser(context.Background(), &phone, &email)
	require.NoError(t, err)
}
