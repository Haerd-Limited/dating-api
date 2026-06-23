package dto

type AccountStatusResponse struct {
	Status            string  `json:"status"`
	SuspendedUntil    *string `json:"suspended_until,omitempty"`
	HasPendingWarning bool    `json:"has_pending_warning"`
}

type ModerationWarningResponse struct {
	ID        string  `json:"id"`
	Message   string  `json:"message"`
	ReportID  *string `json:"report_id,omitempty"`
	CreatedAt string  `json:"created_at"`
}
