package domain

import "time"

const (
	ActionWarnUser    = "warn_user"
	ActionSuspendUser = "suspend_user"
	ActionBanUser     = "ban_user"
	ActionNoAction    = "no_action"
	ActionEscalate    = "escalate"
)

type ModerationWarning struct {
	ID             string
	UserID         string
	ReportID       *string
	Message        string
	CreatedAt      time.Time
	AcknowledgedAt *time.Time
}

type AccountStatusSummary struct {
	Status            string
	SuspendedUntil    *time.Time
	HasPendingWarning bool
}
