package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
)

func SwipeToEntity(s domain.Swipe) entity.Swipe {
	return entity.Swipe{
		ActorID:        s.UserID,
		TargetID:       s.TargetUserID,
		Action:         s.Action,
		PromptID:       null.Int64FromPtr(s.PromptID),
		Message:        null.StringFromPtr(s.Message),
		MessageType:    null.StringFromPtr(s.MessageType),
		VoicenoteURL:   null.StringFromPtr(s.VoiceNoteURL),
		IdempotencyKey: null.StringFromPtr(s.IdempotencyKey),
	}
}
