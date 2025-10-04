package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
)

func SwipesRequestToDomain(swipesReq dto.SwipesRequest, userID string) domain.Swipe {
	return domain.Swipe{
		TargetUserID:   swipesReq.TargetUserID,
		Action:         swipesReq.Action,
		UserID:         userID,
		IdempotencyKey: swipesReq.IdempotencyKey,
		Message:        swipesReq.Message,
		MessageType:    swipesReq.MessageType,
	}
}
