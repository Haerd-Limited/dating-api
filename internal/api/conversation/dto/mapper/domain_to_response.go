package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
)

func MapToGetConversationMessagesResponse(domainMessages []domain.Message) dto.GetConversationMessagesResponse {
	if domainMessages == nil {
		return dto.GetConversationMessagesResponse{
			Messages: []dto.Message{},
		}
	}

	var messages []dto.Message
	for _, msg := range domainMessages {
		messages = append(messages, MapMessageToDto(&msg))
	}

	return dto.GetConversationMessagesResponse{
		Messages: messages,
	}
}

func MapToSendMessageResponse(domainMessage *domain.Message) dto.SendMessageResponse {
	if domainMessage == nil {
		return dto.SendMessageResponse{}
	}
	return dto.SendMessageResponse{
		Messages: MapMessageToDto(domainMessage),
	}
}
func MapMessageToDto(msg *domain.Message) dto.Message {
	if msg == nil {
		return dto.Message{}
	}

	var likedVoicePrompt *dto.VoicePrompt
	if msg.LikedPrompt != nil {
		likedVoicePrompt = &dto.VoicePrompt{
			ID:            msg.LikedPrompt.ID,
			Prompt:        msg.LikedPrompt.Prompt,
			CoverPhotoURL: msg.LikedPrompt.CoverPhotoURL,
			VoiceNoteURL:  msg.LikedPrompt.VoiceNoteURL,
		}
	}

	var snapshot *dto.ScoreSnapshot
	if msg.ResultingScoreSnapShot != nil {
		snapshot = &dto.ScoreSnapshot{
			Threshold: msg.ResultingScoreSnapShot.Threshold,
			Me:        msg.ResultingScoreSnapShot.Me,
			Them:      msg.ResultingScoreSnapShot.Them,
			Revealed:  msg.ResultingScoreSnapShot.Revealed,
			CanReveal: msg.ResultingScoreSnapShot.CanReveal,
			Shared:    msg.ResultingScoreSnapShot.Shared,
		}
	}

	return dto.Message{
		ID:                     msg.ID,
		ConversationID:         msg.ConversationID,
		SenderID:               msg.SenderID,
		Type:                   string(msg.Type),
		TextBody:               msg.TextBody,
		MediaKey:               msg.MediaKey,
		MediaSeconds:           msg.MediaSeconds,
		CreatedAt:              msg.CreatedAt,
		ClientMsgID:            msg.ClientMsgID,
		IsFirstMessage:         msg.IsFirstMessage,
		LikedPrompt:            likedVoicePrompt,
		ResultingScoreSnapShot: snapshot,
	}
}

func MapToGetConversationScoreResponse(score int) dto.GetConversationScoreResponse {
	return dto.GetConversationScoreResponse{
		Score: score,
	}
}

func MapToGetConversationsResponse(conversations []domain.Conversation) dto.GetConversationsResponse {
	if conversations == nil {
		return dto.GetConversationsResponse{
			Conversations: []dto.Conversation{},
		}
	}

	var dtos []dto.Conversation
	for _, convo := range conversations {
		dtos = append(dtos, MapConversationToDto(convo))
	}

	return dto.GetConversationsResponse{
		Conversations: dtos,
	}
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
		Score: dto.ScoreSnapshot{
			Threshold: convo.Score.Threshold,
			Me:        convo.Score.Me,
			Them:      convo.Score.Them,
			Revealed:  convo.Score.Revealed,
			CanReveal: convo.Score.CanReveal,
			Shared:    convo.Score.Shared,
		},
	}
}
