package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
)

func SwipeToEntity(s domain.Swipe) entity.Swipe {
	var promptID null.Int64
	// Treat 0 or nil as invalid/null since 0 is not a valid prompt_id
	if s.PromptID != nil && *s.PromptID != 0 {
		promptID = null.Int64From(*s.PromptID)
	}

	return entity.Swipe{
		ActorID:        s.UserID,
		TargetID:       s.TargetUserID,
		Action:         s.Action,
		PromptID:       promptID,
		Message:        null.StringFromPtr(s.Message),
		MessageType:    null.StringFromPtr(s.MessageType),
		VoicenoteURL:   null.StringFromPtr(s.VoiceNoteURL),
		IdempotencyKey: null.StringFromPtr(s.IdempotencyKey),
	}
}
