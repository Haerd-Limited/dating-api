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
