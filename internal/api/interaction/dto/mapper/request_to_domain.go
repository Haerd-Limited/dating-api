package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/user/dto"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
)

func SwipesRequestToDomain(swipesReq dto.SwipesRequest, userID string) domain.Swipe {
	return domain.Swipe{
		TargetUserID:   swipesReq.TargetUserID,
		Action:         swipesReq.Action,
		UserID:         userID,
		IdempotencyKey: swipesReq.IdempotencyKey,
	}
}
