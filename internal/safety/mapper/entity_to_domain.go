package mapper

import (
	"encoding/json"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
)

func ReportEntityToDomain(report *entity.UserReport) (safetydomain.Report, error) {
	if report == nil {
		return safetydomain.Report{}, nil
	}

	out := safetydomain.Report{
		ID:             report.ID,
		ReporterUserID: report.ReporterUserID,
		ReportedUserID: report.ReportedUserID,
		SubjectType:    safetydomain.SubjectType(report.SubjectType),
		Category:       report.Category,
		Status:         safetydomain.ReportStatus(report.Status),
		CreatedAt:      report.CreatedAt,
		UpdatedAt:      report.UpdatedAt,
	}

	if report.SubjectID.Valid {
		subjectID := report.SubjectID.String
		out.SubjectID = &subjectID
	}

	if report.Notes.Valid {
		notes := report.Notes.String
		out.Notes = &notes
	}

	if report.Severity.Valid {
		severity := report.Severity.String
		out.Severity = &severity
	}

	if report.Metadata.Valid {
		var metadata map[string]any
		if err := json.Unmarshal(report.Metadata.JSON, &metadata); err != nil {
			return safetydomain.Report{}, err
		}

		out.Metadata = metadata
	}

	if report.AutoAction.Valid {
		auto := report.AutoAction.String
		out.AutoAction = &auto
	}

	if report.ResolvedAt.Valid {
		resolved := report.ResolvedAt.Time
		out.ResolvedAt = &resolved
	}

	if report.R != nil {
		for _, action := range report.R.GetReportReportActions() {
			actionDomain, err := ReportActionEntityToDomain(action)
			if err != nil {
				return safetydomain.Report{}, err
			}

			out.Actions = append(out.Actions, actionDomain)
		}
	}

	return out, nil
}

func ReportEntitiesToDomain(reports []*entity.UserReport) ([]safetydomain.Report, error) {
	if len(reports) == 0 {
		return []safetydomain.Report{}, nil
	}

	out := make([]safetydomain.Report, 0, len(reports))

	for _, report := range reports {
		mapped, err := ReportEntityToDomain(report)
		if err != nil {
			return nil, err
		}

		out = append(out, mapped)
	}

	return out, nil
}

func ReportActionEntityToDomain(action *entity.ReportAction) (safetydomain.ReportAction, error) {
	if action == nil {
		return safetydomain.ReportAction{}, nil
	}

	out := safetydomain.ReportAction{
		ID:         action.ID,
		ReportID:   action.ReportID,
		ActionType: action.ActionType,
		CreatedAt:  action.CreatedAt,
	}

	if action.ReviewerID.Valid {
		reviewerID := action.ReviewerID.String
		out.ReviewerID = &reviewerID
	}

	if action.Notes.Valid {
		notes := action.Notes.String
		out.Notes = &notes
	}

	if action.ActionPayload.Valid {
		var payload map[string]any
		if err := json.Unmarshal(action.ActionPayload.JSON, &payload); err != nil {
			return safetydomain.ReportAction{}, err
		}

		out.ActionPayload = payload
	}

	return out, nil
}
