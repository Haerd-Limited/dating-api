package dto

type BlockResponse struct {
	TargetUserID string `json:"target_user_id"`
	Status       string `json:"status"`
}

type ReportResponse struct {
	ID             string            `json:"id"`
	ReporterUserID string            `json:"reporter_user_id"`
	ReportedUserID string            `json:"reported_user_id"`
	SubjectType    string            `json:"subject_type"`
	SubjectID      *string           `json:"subject_id,omitempty"`
	Category       string            `json:"category"`
	Notes          *string           `json:"notes,omitempty"`
	Status         string            `json:"status"`
	Severity       *string           `json:"severity,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
	AutoAction     *string           `json:"auto_action,omitempty"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
	ResolvedAt     *string           `json:"resolved_at,omitempty"`
	Actions        []ReportActionDTO `json:"actions,omitempty"`
}

type ReportActionDTO struct {
	ID         string         `json:"id"`
	ActionType string         `json:"action_type"`
	ActionData map[string]any `json:"action_data,omitempty"`
	Notes      *string        `json:"notes,omitempty"`
	ReviewerID *string        `json:"reviewer_id,omitempty"`
	CreatedAt  string         `json:"created_at"`
}
