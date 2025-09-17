package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
)

func MapMessageToDto(msg *domain.Message) dto.Message {
	if msg == nil {
		return dto.Message{}
	}

	return dto.Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           string(msg.Type),
		TextBody:       msg.TextBody,
		MediaKey:       msg.MediaKey,
		MediaSeconds:   msg.MediaSeconds,
		CreatedAt:      msg.CreatedAt,
		ClientMsgID:    msg.ClientMsgID,
	}
}

func MapConversationsToDtos(conversations []domain.Conversation) []dto.Conversation {
	if conversations == nil {
		return []dto.Conversation{}
	}

	var dtos []dto.Conversation
	for _, convo := range conversations {
		dtos = append(dtos, MapConversationToDto(convo))
	}

	return dtos
}

func MapConversationToDto(convo domain.Conversation) dto.Conversation {
	var message *dto.Message
	if convo.LastMessage != nil {
		message = &dto.Message{
			ID:             convo.LastMessage.ID,
			ConversationID: convo.LastMessage.ConversationID,
			SenderID:       convo.LastMessage.SenderID,
			Type:           string(convo.LastMessage.Type),
			TextBody:       convo.LastMessage.TextBody,
			MediaKey:       convo.LastMessage.MediaKey,
			MediaSeconds:   convo.LastMessage.MediaSeconds,
			CreatedAt:      convo.LastMessage.CreatedAt,
			ClientMsgID:    convo.LastMessage.ClientMsgID,
		}
	}

	return dto.Conversation{
		ID: convo.ID,
		Match: dto.Match{
			ID:          convo.MatchedUser.ID,
			DisplayName: convo.MatchedUser.DisplayName,
			Emoji:       convo.MatchedUser.Emoji,
		},
		CreatedAt:      convo.CreatedAt,
		LastActivityAt: convo.LastActivityAt,
		LastMessage:    message,
		UnreadCount:    convo.UnreadCount,
	}
}
