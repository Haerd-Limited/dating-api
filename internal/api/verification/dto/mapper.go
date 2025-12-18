package dto

import (
	"github.com/Haerd-Limited/dating-api/internal/verification/domain"
)

func MapToStartResponse(dto domain.StartResult) StartResponse {
	return StartResponse{
		SessionID: dto.SessionID,
		Region:    dto.Region,
	}
}

func MapToCompleteResponse(dto domain.CompleteResult) CompleteResponse {
	return CompleteResponse{
		Status:        dto.Status,
		MatchScore:    dto.MatchScore,
		Reasons:       dto.Reasons,
		PhotoVerified: dto.PhotoVerified,
	}
}

func MapToStartVideoResponse(d domain.StartVideoResult) StartVideoResponse {
	return StartVideoResponse{
		Code:      d.Code,
		UploadURL: d.UploadURL,
		UploadKey: d.UploadKey,
	}
}

func MapToSubmitVideoResponse(d domain.SubmitVideoResult) SubmitVideoResponse {
	return SubmitVideoResponse{
		Status: d.Status,
	}
}

func DomainToVideoAttemptResponse(d domain.VideoAttempt) VideoAttemptResponse {
	return VideoAttemptResponse{
		ID:               d.ID,
		UserID:           d.UserID,
		VerificationCode: d.VerificationCode,
		VideoS3Key:       d.VideoS3Key,
		Status:           d.Status,
		RejectionReason:  d.RejectionReason,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func DomainToVideoAttemptListResponse(attempts []domain.VideoAttempt) []VideoAttemptResponse {
	result := make([]VideoAttemptResponse, 0, len(attempts))
	for _, attempt := range attempts {
		result = append(result, DomainToVideoAttemptResponse(attempt))
	}

	return result
}

func ApproveVideoRequestToDomain(req ApproveVideoRequest) domain.ApproveVideoRequest {
	return domain.ApproveVideoRequest{
		Notes: req.Notes,
	}
}

func RejectVideoRequestToDomain(req RejectVideoRequest) domain.RejectVideoRequest {
	return domain.RejectVideoRequest{
		RejectionReason: req.RejectionReason,
		Notes:           req.Notes,
	}
}
