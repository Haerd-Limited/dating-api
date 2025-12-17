package dto

import "github.com/Haerd-Limited/dating-api/internal/verification/domain"

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
