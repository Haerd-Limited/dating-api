package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

func MapMessageDomainToEntity(msg domain.Message) entity.Message {
	return entity.Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           string(msg.Type),
		TextBody:       null.StringFromPtr(msg.TextBody),
		MediaKey:       null.StringFromPtr(msg.MediaUrl),
		MediaSeconds:   utils.FloatPtrToNullDecimal(msg.MediaSeconds),
		ClientMSGID:    null.StringFrom(msg.ClientMsgID),
	}
}
