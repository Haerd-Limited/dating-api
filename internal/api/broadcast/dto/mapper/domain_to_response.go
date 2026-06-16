package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/broadcast/dto"
	"github.com/Haerd-Limited/dating-api/internal/broadcast/domain"
)

func WaitlistUsersToResponse(users []domain.WaitlistUser) []dto.WaitlistUserResponse {
	responses := make([]dto.WaitlistUserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, dto.WaitlistUserResponse{
			ID:             user.ID,
			FirstName:      user.FirstName,
			Phone:          user.Phone,
			OnboardingStep: user.OnboardingStep,
			CreatedAt:      user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			Contacted:      user.Contacted,
		})
	}

	return responses
}

func BroadcastResultToResponse(result domain.BroadcastResult) dto.BroadcastResultResponse {
	recipients := make([]dto.RecipientResultResponse, 0, len(result.Recipients))
	for _, recipient := range result.Recipients {
		recipients = append(recipients, dto.RecipientResultResponse{
			UserID: recipient.UserID,
			Phone:  recipient.Phone,
			Status: recipient.Status,
			Error:  recipient.Error,
		})
	}

	return dto.BroadcastResultResponse{
		Total:      result.Total,
		Sent:       result.Sent,
		Failed:     result.Failed,
		Recipients: recipients,
	}
}
