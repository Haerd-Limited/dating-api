package dto

import "github.com/go-playground/validator/v10"

type BlockRequest struct {
	TargetUserID string  `json:"target_user_id" validate:"required"`
	Reason       *string `json:"reason"`
}

func (br BlockRequest) Validate() error {
	return validator.New().Struct(br)
}

type ReportRequest struct {
	ReportedUserID string         `json:"reported_user_id" validate:"required"`
	SubjectType    string         `json:"subject_type" validate:"omitempty,oneof=user message profile"`
	SubjectID      *string        `json:"subject_id,omitempty"`
	Category       string         `json:"category" validate:"required"`
	Notes          *string        `json:"notes"`
	Severity       *string        `json:"severity"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

func (rr ReportRequest) Validate() error {
	return validator.New().Struct(rr)
}

type ResolveReportRequest struct {
	ActionType string         `json:"action_type" validate:"required"`
	ActionData map[string]any `json:"action_data,omitempty"`
	Notes      *string        `json:"notes"`
	NewStatus  string         `json:"new_status" validate:"required,oneof=pending in_review resolved escalated dismissed"`
	AutoAction *string        `json:"auto_action"`
	ResolvedAt *string        `json:"resolved_at,omitempty"`
}

func (rr ResolveReportRequest) Validate() error {
	return validator.New().Struct(rr)
}
