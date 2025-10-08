package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/mapper"
	scoredomain "github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
)

func (s *service) getConversationByUserIds(ctx context.Context, userID, matchID string) (*domain.Conversation, error) {
	conversationEntity, err := s.conversationRepo.GetConversationByUserIDs(ctx, userID, matchID)
	if err != nil {
		return nil, fmt.Errorf("get conversation by user IDs: %w", err)
	}

	if conversationEntity == nil {
		return nil, nil
	}

	// todo: might be overkill. might be better to just make a get profile/displayname repo method
	matchProfile, err := s.profileService.GetProfileCard(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("get match profile card: %w", err)
	}

	var lastMessage *domain.Message

	if conversationEntity.LastMessageID.Valid {
		var lmErr error

		lastMessage, lmErr = s.getLastMessageByID(ctx, userID, matchID, conversationEntity.LastMessageID.Int64)
		if lmErr != nil {
			return nil, fmt.Errorf("get last message userID=%s matchID=%s: %w", userID, matchID, lmErr)
		}
	} else {
		systemMsg := fmt.Sprintf("Start the chat with %s", matchProfile.DisplayName)
		lastMessage = &domain.Message{
			ConversationID: conversationEntity.ID,
			Type:           domain.MessageTypeSystem,
			TextBody:       &systemMsg,
			CreatedAt:      conversationEntity.CreatedAt,
		}
	}

	var snapShot scoredomain.ScoreSnapshot

	snapShot, err = s.scoreService.GetSnapshot(ctx, conversationEntity.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("get score snapshot: %w", err)
	}

	scoreSnapShot := mapper.MapScoreDomainSnapShotToConversationDomain(snapShot)

	return &domain.Conversation{
		ID: conversationEntity.ID,
		MatchedUser: domain.MatchedUser{
			ID:          matchID,
			DisplayName: matchProfile.DisplayName,
			Emoji:       matchProfile.Emoji,
		},
		CreatedAt:      conversationEntity.CreatedAt,
		LastActivityAt: conversationEntity.LastActivityAt,
		LastMessage:    lastMessage,
		Score:          *scoreSnapShot,
	}, nil
}

func (s *service) getLikedVoicePromptByConvoID(ctx context.Context, convoID string, userID string) (*domain.VoicePrompt, error) {
	firstLike, err := s.getFirstLikeByConvoID(ctx, convoID)
	if err != nil {
		return nil, fmt.Errorf("get first like by convo id userID=%s convoID=%s: %w", userID, convoID, err)
	}

	vp, err := s.profileService.GetVoicePromptByID(ctx, *firstLike.PromptID)
	if err != nil {
		return nil, fmt.Errorf("get voice prompt by id userID=%s convoID=%s swipeID=%v: %w", userID, convoID, firstLike.ID, err)
	}

	return &domain.VoicePrompt{
		ID:            vp.PromptID,
		Prompt:        vp.Prompt,
		CoverPhotoURL: vp.CoverPhotoUrl,
		VoiceNoteURL:  vp.VoiceNoteURL,
	}, nil
}

func (s *service) getFirstLikeByConvoID(ctx context.Context, convoID string) (domain.Swipe, error) {
	convo, err := s.conversationRepo.GetConversationByID(ctx, convoID)
	if err != nil {
		return domain.Swipe{}, fmt.Errorf("get conversation entity by id convoID=%s: %w", convoID, err)
	}

	firstLike, err := s.interactionRepo.GetFirstLikeSwipeByBetweenUsers(ctx, convo.UserA, convo.UserB)
	if err != nil {
		return domain.Swipe{}, fmt.Errorf("get first like swipe by between users convoID=%s: %w", convoID, err)
	}

	if !firstLike.PromptID.Valid {
		return domain.Swipe{}, ErrFirstMessageMissingPromptID
	}

	return mapper.MapSwipeToDomain(firstLike), nil
}

func (s *service) getLastMessageByID(ctx context.Context, userID string, matchID string, lastMessageID int64) (*domain.Message, error) {
	lastMessageEntity, err := s.conversationRepo.GetLastMessageByID(ctx, lastMessageID)
	if err != nil {
		return &domain.Message{}, fmt.Errorf("get last message entity userID=%s matchID=%s: %w", userID, matchID, err)
	}

	if lastMessageEntity == nil {
		return nil, nil
	}

	msg, err := mapper.MapMessageEntityToDomain(*lastMessageEntity)
	if err != nil {
		return &domain.Message{}, fmt.Errorf("map message entity userID=%s matchID=%s: %w", userID, matchID, err)
	}

	return &msg, nil
}

func (s *service) getMatches(ctx context.Context, userID string) ([]domain.Match, error) {
	matchEntities, err := s.conversationRepo.GetMatches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get matches userID=%s: %w", userID, err)
	}

	if len(matchEntities) == 0 {
		return []domain.Match{}, nil
	}

	return mapper.MapMatchEntitiesToDomain(matchEntities), nil
}

func (s *service) sendMessageToConversation(msg domain.Message) {
	if msg.Type == domain.MessageTypeSystem {
		return
	}
	// Build server event (use your DTO for payload)
	evt := dto.ServerMsg{
		Type:           "message.new",
		ConversationID: msg.ConversationID,
		Payload: map[string]any{
			"id":              msg.ID,
			"conversation_id": msg.ConversationID,
			"sender_id":       msg.SenderID,
			"text_body":       msg.TextBody,
			"type":            msg.Type,
			"media_key":       msg.MediaKey,
			"media_seconds":   msg.MediaSeconds,
			"created_at":      msg.CreatedAt,
			"client_msg_id":   msg.ClientMsgID,
		},
	}
	b, _ := json.Marshal(evt)
	s.hub.BroadcastToConversation(msg.ConversationID, b)
}
