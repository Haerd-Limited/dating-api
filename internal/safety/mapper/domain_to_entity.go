package mapper

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
)

func BlockRequestToEntity(req safetydomain.BlockRequest) *entity.UserBlock {
	out := entity.UserBlock{
		BlockerUserID: req.BlockerID,
		BlockedUserID: req.BlockedID,
		CreatedAt:     time.Now().UTC(),
	}

	if reason := normalizeStringPointer(req.Reason); reason != nil {
		out.Reason = null.StringFrom(*reason)
	}

	return &out
}

func ReportRequestToEntity(req safetydomain.ReportRequest) (*entity.UserReport, error) {
	now := time.Now().UTC()

	out := entity.UserReport{
		ReporterUserID: req.ReporterUserID,
		ReportedUserID: req.ReportedUserID,
		SubjectType:    string(req.SubjectType),
		Category:       req.Category,
		Status:         string(safetydomain.ReportStatusPending),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if subjectID := normalizeStringPointer(req.SubjectID); subjectID != nil {
		out.SubjectID = null.StringFrom(*subjectID)
	}

	if notes := normalizeStringPointer(req.Notes); notes != nil {
		out.Notes = null.StringFrom(*notes)
	}

	if severity := normalizeStringPointer(req.Severity); severity != nil {
		out.Severity = null.StringFrom(*severity)
	}

	if len(req.Metadata) > 0 {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, err
		}

		out.Metadata = null.JSONFrom(metadataBytes)
	}

	return &out, nil
}

func ResolveRequestToActionEntity(req safetydomain.ResolveReportRequest) (*entity.ReportAction, error) {
	out := entity.ReportAction{
		ReportID:   req.ReportID,
		ActionType: req.ActionType,
		CreatedAt:  time.Now().UTC(),
	}

	if req.ReviewerID != "" {
		out.ReviewerID = null.StringFrom(req.ReviewerID)
	}

	if req.Notes != nil {
		if notes := strings.TrimSpace(*req.Notes); notes != "" {
			out.Notes = null.StringFrom(notes)
		}
	}

	if len(req.ActionData) > 0 {
		payloadBytes, err := json.Marshal(req.ActionData)
		if err != nil {
			return nil, err
		}

		out.ActionPayload = null.JSONFrom(payloadBytes)
	}

	return &out, nil
}

func ApplyResolutionToReportEntity(report *entity.UserReport, req safetydomain.ResolveReportRequest) {
	report.Status = string(req.NewStatus)
	report.UpdatedAt = time.Now().UTC()

	if autoAction := normalizeStringPointer(req.AutoAction); autoAction != nil {
		report.AutoAction = null.StringFrom(*autoAction)
	} else {
		report.AutoAction = null.String{}
	}

	if req.ResolvedAt != nil {
		report.ResolvedAt = null.TimeFrom(*req.ResolvedAt)
	} else {
		report.ResolvedAt = null.Time{}
	}
}

func normalizeStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
