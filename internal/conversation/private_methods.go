package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	dtomapper "github.com/Haerd-Limited/dating-api/internal/api/conversation/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/mapper"
	scoredomain "github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
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
			Theme: domain.Theme{
				BaseHex: matchProfile.Theme.BaseHex,
				Palette: matchProfile.Theme.Palette,
			},
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
		Payload:        dtomapper.MapMessageToDto(&msg),
	}
	b, _ := json.Marshal(evt)
	s.hub.BroadcastToConversation(msg.ConversationID, b)
}

func (s *service) updateConversationRealtime(ctx context.Context, userID string, convoID string) {
	participants, err := s.conversationRepo.GetConversationParticipants(ctx, convoID)
	if err != nil {
		s.logger.Error("updating conversation: get conversation participants", zap.Error(err))
		return
	}

	if participants == nil {
		s.logger.Error("updating conversation: participants is nil")
		return
	}

	if len(participants) != 2 {
		s.logger.Error("updating conversation: participants is not 2")
		return
	}

	if participants[0].UserID == participants[1].UserID {
		s.logger.Error("updating conversation: participants are the same")
		return
	}

	var matchID string
	if participants[0].UserID == userID {
		matchID = participants[1].UserID
	} else {
		matchID = participants[0].UserID
	}
	// from user POV
	convo, err := s.getConversationByUserIds(ctx, userID, matchID)
	if err != nil {
		s.logger.Error("updating conversation: get conversation by user ids", zap.Error(err))
		return
	}

	if convo == nil {
		s.logger.Error("updating conversation: conversation is nil")
		return
	}

	// from user POV
	if convo.LastMessage == nil {
		s.logger.Error("updating conversation: last message is nil")
		return
	}

	evt := dto.Event{
		ID:        realtime.NewEventID(),
		Type:      "conversation.updated",
		ActorID:   convo.LastMessage.SenderID,
		Ts:        time.Now().UTC(),
		ContextID: convo.ID,
		Data:      dtomapper.MapConversationToDto(*convo),
		Version:   1,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("error marshalling event", zap.Error(err))
		return
	}

	s.hub.BroadcastToUser(userID, b)

	// from match POV
	convo, err = s.getConversationByUserIds(ctx, matchID, userID)
	if err != nil {
		s.logger.Error("updating conversation from match POV: get conversation by user ids", zap.Error(err))
		return
	}

	if convo == nil {
		s.logger.Error("updating conversation from match POV: conversation is nil")
		return
	}

	if convo.LastMessage == nil {
		s.logger.Error("updating conversation: last message is nil")
		return
	}

	evt = dto.Event{
		ID:        realtime.NewEventID(),
		Type:      "conversation.updated",
		ActorID:   convo.LastMessage.SenderID,
		Ts:        time.Now().UTC(),
		ContextID: convo.ID,
		Data:      dtomapper.MapConversationToDto(*convo),
		Version:   1,
	}

	byts, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("error marshalling event", zap.Error(err))
		return
	}

	s.hub.BroadcastToUser(matchID, byts)
}
