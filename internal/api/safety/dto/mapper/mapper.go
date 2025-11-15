package mapper

import (
	"fmt"
	"strings"
	"time"

	dto "github.com/Haerd-Limited/dating-api/internal/api/safety/dto"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
)

func BlockRequestToDomain(req dto.BlockRequest, blockerID string) safetydomain.BlockRequest {
	return safetydomain.BlockRequest{
		BlockerID: blockerID,
		BlockedID: req.TargetUserID,
		Reason:    req.Reason,
	}
}

func ReportRequestToDomain(req dto.ReportRequest, reporterID string) safetydomain.ReportRequest {
	return safetydomain.ReportRequest{
		ReporterUserID: reporterID,
		ReportedUserID: req.ReportedUserID,
		SubjectType:    safetydomain.SubjectType(req.SubjectType),
		SubjectID:      req.SubjectID,
		Category:       req.Category,
		Notes:          req.Notes,
		Severity:       req.Severity,
		Metadata:       req.Metadata,
	}
}

func ResolveReportRequestToDomain(req dto.ResolveReportRequest, reportID, reviewerID string) (safetydomain.ResolveReportRequest, error) {
	var resolvedAt *time.Time

	if req.ResolvedAt != nil {
		trimmed := strings.TrimSpace(*req.ResolvedAt)
		if trimmed != "" {
			parsed, err := time.Parse(time.RFC3339, trimmed)
			if err != nil {
				return safetydomain.ResolveReportRequest{}, fmt.Errorf("invalid resolved_at format: %w", err)
			}

			resolvedAt = &parsed
		}
	}

	return safetydomain.ResolveReportRequest{
		ReportID:   reportID,
		ReviewerID: reviewerID,
		ActionType: req.ActionType,
		ActionData: req.ActionData,
		Notes:      req.Notes,
		NewStatus:  safetydomain.ReportStatus(req.NewStatus),
		AutoAction: req.AutoAction,
		ResolvedAt: resolvedAt,
	}, nil
}

func MapReportDomainToDTO(report safetydomain.Report) dto.ReportResponse {
	resp := dto.ReportResponse{
		ID:             report.ID,
		ReporterUserID: report.ReporterUserID,
		ReportedUserID: report.ReportedUserID,
		SubjectType:    string(report.SubjectType),
		Category:       report.Category,
		Status:         string(report.Status),
		CreatedAt:      report.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      report.UpdatedAt.UTC().Format(time.RFC3339),
		Metadata:       report.Metadata,
		AutoAction:     report.AutoAction,
		Severity:       report.Severity,
		Notes:          report.Notes,
		SubjectID:      report.SubjectID,
	}

	if report.ResolvedAt != nil {
		resolved := report.ResolvedAt.UTC().Format(time.RFC3339)
		resp.ResolvedAt = &resolved
	}

	if len(report.Actions) > 0 {
		resp.Actions = make([]dto.ReportActionDTO, 0, len(report.Actions))

		for _, action := range report.Actions {
			actionDTO := dto.ReportActionDTO{
				ID:         action.ID,
				ActionType: action.ActionType,
				ActionData: action.ActionPayload,
				ReviewerID: action.ReviewerID,
				Notes:      action.Notes,
				CreatedAt:  action.CreatedAt.UTC().Format(time.RFC3339),
			}

			resp.Actions = append(resp.Actions, actionDTO)
		}
	}

	return resp
}

func MapReportsDomainToDTO(reports []safetydomain.Report) []dto.ReportResponse {
	out := make([]dto.ReportResponse, 0, len(reports))
	for _, report := range reports {
		out = append(out, MapReportDomainToDTO(report))
	}

	return out
}
