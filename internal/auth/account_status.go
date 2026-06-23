package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

var (
	ErrAccountBanned    = errors.New("account banned")
	ErrAccountSuspended = errors.New("account suspended")
)

type SuspendedAccountError struct {
	Until time.Time
}

func (e *SuspendedAccountError) Error() string {
	return ErrAccountSuspended.Error()
}

func (e *SuspendedAccountError) Unwrap() error {
	return ErrAccountSuspended
}

func (as *authService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if err := as.AuthRepo.RevokeAllRefreshTokens(ctx, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens userID=%s: %w", userID, err)
	}

	return nil
}

func (as *authService) ensureAccountCanAuthenticate(ctx context.Context, userID string) error {
	state, err := as.UserService.GetAccountGateState(ctx, userID)
	if err != nil {
		return err
	}

	effective := state.EffectiveStatus(time.Now().UTC())

	switch effective {
	case userdomain.AccountStatusBanned:
		return ErrAccountBanned
	case userdomain.AccountStatusSuspended:
		if state.SuspendedUntil == nil {
			return ErrAccountSuspended
		}

		return &SuspendedAccountError{Until: *state.SuspendedUntil}
	default:
		return nil
	}
}
