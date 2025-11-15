package domain

import "time"

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusInReview  ReportStatus = "in_review"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusEscalated ReportStatus = "escalated"
	ReportStatusDismissed ReportStatus = "dismissed"
)

type SubjectType string

const (
	SubjectTypeUser    SubjectType = "user"
	SubjectTypeMessage SubjectType = "message"
	SubjectTypeProfile SubjectType = "profile"
)

type Report struct {
	ID             string
	ReporterUserID string
	ReportedUserID string
	SubjectType    SubjectType
	SubjectID      *string
	Category       string
	Notes          *string
	Status         ReportStatus
	Severity       *string
	Metadata       map[string]any
	AutoAction     *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ResolvedAt     *time.Time
	Actions        []ReportAction
}

type ReportAction struct {
	ID            string
	ReportID      string
	ReviewerID    *string
	ActionType    string
	ActionPayload map[string]any
	Notes         *string
	CreatedAt     time.Time
}

type ReportRequest struct {
	ReporterUserID string
	ReportedUserID string
	SubjectType    SubjectType
	SubjectID      *string
	Category       string
	Notes          *string
	Severity       *string
	Metadata       map[string]any
}

type ReportListFilter struct {
	Status   []ReportStatus
	Category []string
	Reporter *string
	Reported *string
	Limit    int
	Offset   int
}

type ResolveReportRequest struct {
	ReportID   string
	ReviewerID string
	ActionType string
	ActionData map[string]any
	Notes      *string
	NewStatus  ReportStatus
	AutoAction *string
	ResolvedAt *time.Time
}
