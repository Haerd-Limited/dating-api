package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func TestEnsureAccountCanAuthenticate(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(24 * time.Hour)

	ctrl := gomock.NewController(t)
	userSvc := user.NewMockService(ctrl)

	as := &authService{
		logger:      zaptest.NewLogger(t),
		UserService: userSvc,
	}

	t.Run("banned", func(t *testing.T) {
		userSvc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
			Status: userdomain.AccountStatusBanned,
		}, nil)

		err := as.ensureAccountCanAuthenticate(context.Background(), "user-1")
		assert.ErrorIs(t, err, ErrAccountBanned)
	})

	t.Run("suspended", func(t *testing.T) {
		userSvc.EXPECT().GetAccountGateState(gomock.Any(), "user-2").Return(userdomain.AccountState{
			Status:         userdomain.AccountStatusSuspended,
			SuspendedUntil: &future,
		}, nil)

		err := as.ensureAccountCanAuthenticate(context.Background(), "user-2")
		require.Error(t, err)

		var suspendedErr *SuspendedAccountError

		assert.ErrorAs(t, err, &suspendedErr)
		assert.Equal(t, future, suspendedErr.Until)
	})

	t.Run("active", func(t *testing.T) {
		userSvc.EXPECT().GetAccountGateState(gomock.Any(), "user-3").Return(userdomain.AccountState{
			Status: userdomain.AccountStatusActive,
		}, nil)

		err := as.ensureAccountCanAuthenticate(context.Background(), "user-3")
		assert.NoError(t, err)
	})
}
