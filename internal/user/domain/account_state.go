package domain

import "time"

const (
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
	AccountStatusBanned    = "banned"
)

type AccountState struct {
	Status            string
	SuspendedUntil    *time.Time
	Reason            *string
	HasPendingWarning bool
}

// EffectiveStatus returns the status used for enforcement (lazy suspension expiry).
func (s AccountState) EffectiveStatus(now time.Time) string {
	if s.Status == AccountStatusSuspended && s.SuspendedUntil != nil && !s.SuspendedUntil.After(now) {
		return AccountStatusActive
	}

	return s.Status
}
