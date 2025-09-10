package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/aarondl/null/v8"
)

func SwipeToEntity(s domain.Swipe) entity.Swipe {
	return entity.Swipe{
		ActorID:        s.UserID,
		TargetID:       s.TargetUserID,
		Action:         s.Action,
		IdempotencyKey: null.StringFromPtr(s.IdempotencyKey),
	}
}
