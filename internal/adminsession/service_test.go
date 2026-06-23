package adminsession

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/adminsession/domain"
)

func TestCreateSessionRejectsNameNotInRoster(t *testing.T) {
	svc := NewService(zaptest.NewLogger(t), nil, []string{"Dhrubo Roy", "Jane Smith"})

	_, err := svc.CreateSession(context.Background(), domain.CreateSessionRequest{
		DisplayName: "Unknown Person",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDisplayName)
}

func TestCreateSessionRejectsEmptyDisplayName(t *testing.T) {
	svc := NewService(zaptest.NewLogger(t), nil, []string{"Dhrubo Roy"})

	_, err := svc.CreateSession(context.Background(), domain.CreateSessionRequest{
		DisplayName: "   ",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDisplayName)
}

func TestParseRoster(t *testing.T) {
	assert.Equal(t, []string{"Alice", "Bob"}, ParseRoster("Alice, Bob"))
	assert.Nil(t, ParseRoster(""))
	assert.Empty(t, ParseRoster("  ,  "))
}
