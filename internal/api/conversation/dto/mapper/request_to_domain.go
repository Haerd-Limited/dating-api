package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
)

func MapSendMessageRequestToDomain(req dto.SendMessageRequest, convoID string, userID string) domain.Message {
	return domain.Message{
		ClientMsgID:    req.ClientMsgID,
		Type:           domain.MessageType(req.Type),
		TextBody:       req.TextBody,
		MediaKey:       req.MediaKey,
		MediaSeconds:   req.MediaSeconds,
		ConversationID: convoID,
		SenderID:       userID,
	}
}
