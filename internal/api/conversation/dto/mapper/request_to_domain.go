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
		MediaUrl:       req.MediaUrl,
		MediaSeconds:   req.MediaSeconds,
		ConversationID: convoID,
		SenderID:       userID,
	}
}

func MapMakeRevealDecisionRequestToDomain(req dto.MakeRevealDecisionRequest) domain.RevealDecision {
	return domain.RevealDecision(req.Decision)
}
