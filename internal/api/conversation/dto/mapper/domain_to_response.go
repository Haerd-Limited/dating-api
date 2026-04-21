package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
)

func MapToGetConversationMessagesResponse(domainMessages []domain.Message, userID string) dto.GetConversationMessagesResponse {
	if domainMessages == nil {
		return dto.GetConversationMessagesResponse{
			Messages: []dto.Message{},
		}
	}

	var messages []dto.Message
	for _, msg := range domainMessages {
		messages = append(messages, MapMessageToDto(&msg, userID))
	}

	return dto.GetConversationMessagesResponse{
		Messages: messages,
	}
}

func MapToSendMessageResponse(domainMessage *domain.Message, userID string) dto.SendMessageResponse {
	if domainMessage == nil {
		return dto.SendMessageResponse{}
	}

	return dto.SendMessageResponse{
		Messages: MapMessageToDto(domainMessage, userID),
	}
}

func MapMessageToDto(msg *domain.Message, userID string) dto.Message {
	if msg == nil {
		return dto.Message{}
	}

	var likedVoicePrompt *dto.VoicePrompt
	if msg.LikedPrompt != nil {
		likedVoicePrompt = &dto.VoicePrompt{
			ID:                    msg.LikedPrompt.ID,
			Prompt:                msg.LikedPrompt.Prompt,
			CoverMediaURL:         msg.LikedPrompt.CoverMediaURL,
			CoverMediaType:        msg.LikedPrompt.CoverMediaType,
			CoverMediaAspectRatio: msg.LikedPrompt.CoverMediaAspectRatio,
			VoiceNoteURL:          msg.LikedPrompt.VoiceNoteURL,
			WaveformData:          msg.LikedPrompt.WaveformData,
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

	textBody := msg.TextBody

	// If the message is a voice note and TextBody is nil, populate it with a summary
	if msg.Type == domain.MessageTypeVoice {
		var summaryText string
		if msg.SenderID == userID {
			summaryText = "You sent a voice note"
		} else {
			summaryText = "You received a voice note"
		}
		textBody = &summaryText
	}

	return dto.Message{
		ID:                     msg.ID,
		ConversationID:         msg.ConversationID,
		SenderID:               msg.SenderID,
		Type:                   string(msg.Type),
		TextBody:               textBody,
		MediaKey:               msg.MediaUrl,
		MediaSeconds:           msg.MediaSeconds,
		CreatedAt:              msg.CreatedAt,
		ClientMsgID:            msg.ClientMsgID,
		IsFirstMessage:         msg.IsFirstMessage,
		LikedPrompt:            likedVoicePrompt,
		ResultingScoreSnapShot: snapshot,
	}
}

func MapToGetConversationsResponse(conversations []domain.Conversation, userID string) dto.GetConversationsResponse {
	if conversations == nil {
		return dto.GetConversationsResponse{
			Conversations: []dto.Conversation{},
		}
	}

	var dtos []dto.Conversation
	for _, convo := range conversations {
		dtos = append(dtos, MapConversationToDto(convo, userID))
	}

	return dto.GetConversationsResponse{
		Conversations: dtos,
	}
}

func MapConversationToDto(convo domain.Conversation, userID string) dto.Conversation {
	var message *dto.Message
	if convo.LastMessage != nil {
		textBody := convo.LastMessage.TextBody

		// If the last message is a voice note and TextBody is nil, populate it with a summary
		if convo.LastMessage.Type == domain.MessageTypeVoice {
			var summaryText string
			if convo.LastMessage.SenderID == userID {
				summaryText = "You sent a voice note"
			} else {
				summaryText = "You received a voice note"
			}
			textBody = &summaryText
		}

		message = &dto.Message{
			ID:             convo.LastMessage.ID,
			ConversationID: convo.LastMessage.ConversationID,
			SenderID:       convo.LastMessage.SenderID,
			Type:           string(convo.LastMessage.Type),
			TextBody:       textBody,
			MediaKey:       convo.LastMessage.MediaUrl,
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
			Theme: dto.UserTheme{
				BaseHex: convo.MatchedUser.Theme.BaseHex,
				Palette: convo.MatchedUser.Theme.Palette,
			},
		},
		CreatedAt:      convo.CreatedAt,
		LastActivityAt: convo.LastActivityAt,
		LastMessage:    message,
		UnreadCount:    convo.UnreadCount,
		DateMode:       convo.DateMode,
		Photos:         MapPhotosToDTO(convo.Photos),
	}
}

func MapPhotosToDTO(photos []domain.Photo) []dto.Photo {
	if photos == nil {
		return nil
	}

	var dtos []dto.Photo
	for _, photo := range photos {
		dtos = append(dtos, dto.Photo{
			URL:       photo.URL,
			IsPrimary: photo.IsPrimary,
			Position:  photo.Position,
		})
	}

	return dtos
}
