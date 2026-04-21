package conversation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/mapper"
	"github.com/Haerd-Limited/dating-api/internal/conversation/score"
	scoredomain "github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	storage3 "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

type Service interface {
	GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error)
	CreateConversation(ctx context.Context, userID, matchUserID string) error
	CreateConversationViaTx(ctx context.Context, userID, matchUserID string, tx *sql.Tx) (string, error)
	SendMessage(ctx context.Context, tx *sql.Tx, msg domain.Message) (domain.Message, error)
	GetMessages(ctx context.Context, convoID string, userID string) ([]domain.Message, error)
	IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error)
	CreateConversationScores(ctx context.Context, convoID, userID, matchedUserID string, tx *sql.Tx) error
	Unmatch(ctx context.Context, userID, conversationID string, reason string) error
}

type service struct {
	logger           *zap.Logger
	conversationRepo storage.ConversationRepository
	profileService   profile.Service
	flake            interface{ Next() int64 }
	hub              realtime.Broadcaster
	interactionRepo  storage3.InteractionRepository
	scoreService     score.Service
	uow              uow.UoW
	notificationSvc  notification.Service
}

func NewConversationService(
	logger *zap.Logger,
	conversationRepo storage.ConversationRepository,
	profileService profile.Service,
	flake interface{ Next() int64 },
	hub realtime.Broadcaster,
	interactionRepo storage3.InteractionRepository,
	scoreService score.Service,
	uow uow.UoW,
	notificationSvc notification.Service,
) Service {
	return &service{
		logger:           logger,
		conversationRepo: conversationRepo,
		profileService:   profileService,
		flake:            flake,
		hub:              hub,
		interactionRepo:  interactionRepo,
		scoreService:     scoreService,
		uow:              uow,
		notificationSvc:  notificationSvc,
	}
}

var (
	ErrFirstMessageMissingPromptID         = errors.New("first message missing prompt id")
	ErrInvalidMessageType                  = errors.New("invalid message type")
	ErrMissingRequiredFieldToSendVoicenote = errors.New("missing required field to send voice note")
	ErrGifMessageMissingURL                = errors.New("gif message missing url")
	ErrVoiceNoteTooLong                    = errors.New("voice note too long")
	ErrInvalidMessage                      = errors.New("invalid message")
	ErrInvalidVoiceNoteSeconds             = errors.New("invalid voice note seconds")
	ErrTextTooLong                         = errors.New("text too long")
	ErrInvalidTextMessage                  = errors.New("invalid text message")
)

const messagePreviewMaxRunes = 120

func (s *service) IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	return s.conversationRepo.IsConversationParticipant(ctx, conversationID, userID)
}

func (s *service) CreateConversationScores(ctx context.Context, convoID, userID, matchedUserID string, tx *sql.Tx) error {
	return s.conversationRepo.CreateConversationScores(ctx, convoID, userID, matchedUserID, tx)
}

func (s *service) GetMessages(ctx context.Context, convoID string, userID string) ([]domain.Message, error) {
	messageEntities, err := s.conversationRepo.GetMessagesByConversationID(ctx, convoID, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get messages", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	// Assuming that both participants simply liked each-other.
	if len(messageEntities) == 0 {
		var firstLike domain.Swipe

		firstLike, err = s.getFirstLikeByConvoID(ctx, convoID)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "get first like by convo id", err, zap.String("userID", userID), zap.String("convoID", convoID))
		}

		_, err = s.SendMessage(ctx, nil, domain.Message{
			ConversationID: convoID,
			SenderID:       firstLike.ActorID,
			Type:           domain.MessageType(*firstLike.MessageType),
			TextBody:       firstLike.Message,
			ClientMsgID:    uuid.New().String(),
		})
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "send message", err, zap.String("userID", userID), zap.String("convoID", convoID))
		}

		messageEntities, err = s.conversationRepo.GetMessagesByConversationID(ctx, convoID, userID)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "get messages", err, zap.String("userID", userID), zap.String("convoID", convoID))
		}
	}

	// determine first sent message
	firstSentMessageIndex := 0
	for i := 1; i < len(messageEntities); i++ {
		if messageEntities[i].CreatedAt.Before(messageEntities[firstSentMessageIndex].CreatedAt) {
			firstSentMessageIndex = i
		}
	}

	var messages []domain.Message

	var msg domain.Message
	for index, messageEntity := range messageEntities {
		msg, err = mapper.MapMessageEntityToDomain(*messageEntity)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "map message entity", err, zap.String("userID", userID), zap.String("convoID", convoID))
		}

		// get prompt for first message
		if index == firstSentMessageIndex {
			msg.LikedPrompt, err = s.getLikedVoicePromptByConvoID(ctx, convoID, userID)
			if err != nil {
				return nil, commonlogger.LogError(s.logger, "get liked prompt by convo id", err, zap.String("userID", userID), zap.String("convoID", convoID))
			}

			msg.IsFirstMessage = true
		}

		messages = append(messages, msg)
	}

	// Mark all messages in the conversation as read for this user
	err = s.conversationRepo.MarkConversationMessagesAsRead(ctx, convoID, userID, nil)
	if err != nil {
		// Log error but don't fail the request - read status is best-effort
		s.logger.Sugar().Warnw("failed to mark conversation messages as read", "error", err, "userID", userID, "convoID", convoID)
	}

	return messages, nil
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	// 1. Get Matches.
	// this is to make sure user's are matched since a user can't have a convo before matching. and if unmatched, convo should be gone
	matches, err := s.getMatches(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get matches", err, zap.String("userID", userID))
	}

	if len(matches) == 0 {
		return []domain.Conversation{}, nil // user has no conversations/matches
	}

	// 2. Based on matched users, for loop and get conversation details
	var conversations []domain.Conversation

	for _, match := range matches {
		var matchUserID string
		if match.UserA == userID {
			matchUserID = match.UserB
		} else {
			matchUserID = match.UserA
		}

		var conversation *domain.Conversation

		conversation, convoErr := s.getConversationByUserIds(ctx, userID, matchUserID)
		if convoErr != nil {
			return nil, commonlogger.LogError(s.logger, "get conversation", convoErr, zap.String("userID", userID), zap.String("matchUserID", matchUserID))
		}

		if conversation == nil {
			s.logger.Warn("convo_self_heal", zap.String("userID", userID), zap.String("matchUserID", matchUserID))
			// create convo
			_, createConvoErr := s.conversationRepo.CreateConversation(ctx, userID, matchUserID, nil)
			if createConvoErr != nil {
				return nil, commonlogger.LogError(s.logger, "create conversation", createConvoErr, zap.String("userID", userID), zap.String("matchUserID", matchUserID))
			}

			conversation, convoErr = s.getConversationByUserIds(ctx, userID, matchUserID)
			if convoErr != nil {
				return nil, commonlogger.LogError(s.logger, "get conversation", convoErr, zap.String("userID", userID), zap.String("matchUserID", matchUserID))
			}
		}

		conversations = append(conversations, *conversation)
	}

	// order by latest match/message
	for i := 0; i < len(conversations); i++ {
		for j := i + 1; j < len(conversations); j++ {
			if conversations[i].LastMessage.CreatedAt.Before(conversations[j].LastMessage.CreatedAt) {
				conversations[i], conversations[j] = conversations[j], conversations[i]
			}
		}
	}

	return conversations, nil
}

func (s *service) CreateConversation(ctx context.Context, userID, matchUserID string) error {
	_, err := s.conversationRepo.CreateConversation(ctx, userID, matchUserID, nil)
	if err != nil {
		return commonlogger.LogError(s.logger, "create conversation", err, zap.String("userID", userID), zap.String("matchUserID", matchUserID))
	}

	return nil
}

func (s *service) CreateConversationViaTx(ctx context.Context, userID, matchUserID string, tx *sql.Tx) (string, error) {
	convoEntity, err := s.conversationRepo.CreateConversation(ctx, userID, matchUserID, tx)
	if err != nil {
		return "", commonlogger.LogError(s.logger, "create conversation via tx", err, zap.String("userID", userID), zap.String("matchUserID", matchUserID))
	}

	return convoEntity.ID, nil
}

func (s *service) validateMessage(msg domain.Message) error {
	// Basic invariants common to all messages
	if msg.ConversationID == "" {
		return fmt.Errorf("%w: missing conversation_id", ErrInvalidMessage)
	}

	if msg.SenderID == "" {
		return fmt.Errorf("%w: missing sender_id", ErrInvalidMessage)
	}

	switch msg.Type {
	case domain.MessageTypeSystem:
		return nil
	case domain.MessageTypeText:
		// Require non-empty text (adjust per your product rules)
		if msg.TextBody == nil || strings.TrimSpace(*msg.TextBody) == "" {
			return fmt.Errorf("%w: empty text body", ErrInvalidTextMessage)
		}
		// Optional: length cap
		if len([]rune(*msg.TextBody)) > constants.MaxTextLengthRunes {
			return fmt.Errorf("%w: text length exceeds %d characters", ErrTextTooLong, constants.MaxTextLengthRunes)
		}

	case domain.MessageTypeVoice:
		// Require BOTH url and seconds
		if msg.MediaUrl == nil || msg.MediaSeconds == nil {
			return fmt.Errorf("%w: voice note requires media_url and media_seconds", ErrMissingRequiredFieldToSendVoicenote)
		}

		secs := *msg.MediaSeconds
		if !(secs > 0) || math.IsNaN(secs) || math.IsInf(secs, 0) {
			return fmt.Errorf("%w: invalid media_seconds=%v", ErrInvalidVoiceNoteSeconds, secs)
		}

		if secs > float64(constants.MaxVoiceNoteLengthInSeconds) {
			return fmt.Errorf("%w: cannot be greater than %v seconds (got %.3f)",
				ErrVoiceNoteTooLong, constants.MaxVoiceNoteLengthInSeconds, secs)
		}

		if err := utils.ValidateHTTPURL(*msg.MediaUrl); err != nil {
			return fmt.Errorf("%w: media_url invalid: %v", commonErrors.ErrInvalidMediaUrl, err)
		}
		// Optional: enforce allowed MIME/container or extension

	case domain.MessageTypeGif:
		if msg.MediaUrl == nil {
			return fmt.Errorf("%w: media_url is required", ErrGifMessageMissingURL)
		}

		if err := utils.ValidateHTTPURL(*msg.MediaUrl); err != nil {
			return fmt.Errorf("%w: media_url invalid: %v", commonErrors.ErrInvalidMediaUrl, err)
		}
		// Optional: allow-list providers (e.g., tenor, giphy)

	default:
		return fmt.Errorf("%w: type=%s", ErrInvalidMessageType, msg.Type)
	}

	return nil
}

func (s *service) SendMessage(ctx context.Context, tx *sql.Tx, msg domain.Message) (domain.Message, error) {
	err := s.validateMessage(msg)
	if err != nil {
		return domain.Message{}, commonlogger.LogError(s.logger, "validate message", err, zap.Any("message", msg))
	}

	msg.ID = s.flake.Next()
	ent := mapper.MapMessageDomainToEntity(msg)

	e, err := s.conversationRepo.SendMessageViaTx(ctx, tx, ent)
	if err != nil {
		return domain.Message{}, commonlogger.LogError(s.logger, "send message via tx", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
	}

	result, err := mapper.MapMessageEntityToDomain(*e)
	if err != nil {
		return domain.Message{}, commonlogger.LogError(s.logger, "map message entity", err, zap.String("userID", msg.SenderID))
	}

	result.ResultingScoreSnapShot, err = s.ApplyScore(ctx, tx, result)
	if err != nil {
		return domain.Message{}, commonlogger.LogError(s.logger, "apply score via tx", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
	}

	s.sendMessageToConversation(result)

	s.updateConversationRealtime(ctx, msg.SenderID, msg.ConversationID)

	s.sendNewMessagePush(ctx, result)

	// analytics: message sent
	props := map[string]any{
		"type":      string(result.Type),
		"has_voice": result.Type == domain.MessageTypeVoice,
		"len": func() int {
			if result.TextBody == nil {
				return 0
			}
			return len([]rune(*result.TextBody))
		}(),
		"client_msg_id": result.ClientMsgID,
	}
	commonanalytics.Track(ctx, "conversation.message_sent", &result.SenderID, nil, props)

	return result, nil
}

func (s *service) sendNewMessagePush(ctx context.Context, msg domain.Message) {
	if s.notificationSvc == nil {
		return
	}

	if msg.Type == domain.MessageTypeSystem {
		return
	}

	participants, err := s.conversationRepo.GetConversationParticipants(ctx, msg.ConversationID)
	if err != nil {
		s.logger.Sugar().Warnw("send new message push: get participants", "error", err, "conversationID", msg.ConversationID)
		return
	}

	var recipientID string

	for _, participant := range participants {
		if participant.UserID != msg.SenderID {
			recipientID = participant.UserID
			break
		}
	}

	if recipientID == "" {
		return
	}

	senderProfile, err := s.profileService.GetProfileCard(ctx, msg.SenderID)
	if err != nil {
		s.logger.Sugar().Warnw("send new message push: get sender profile", "error", err, "senderID", msg.SenderID)
		return
	}

	preview := buildMessagePreview(msg, senderProfile.DisplayName)
	if preview == "" {
		preview = "You have a new message waiting for you."
	}

	if err := s.notificationSvc.SendNewMessageNotification(ctx, msg.SenderID, senderProfile.DisplayName, msg.ConversationID, recipientID, preview); err != nil {
		s.logger.Sugar().Warnw("failed to send new message notification", "error", err, "conversationID", msg.ConversationID, "recipientID", recipientID)
	}
}

func buildMessagePreview(msg domain.Message, senderName string) string {
	// Extract first name from sender name (split by space and take first part)
	firstName := senderName
	if spaceIdx := strings.Index(senderName, " "); spaceIdx > 0 {
		firstName = senderName[:spaceIdx]
	}

	switch msg.Type {
	case domain.MessageTypeText:
		if msg.TextBody == nil {
			return ""
		}

		trimmed := strings.TrimSpace(*msg.TextBody)
		if trimmed == "" {
			return ""
		}

		runes := []rune(trimmed)
		if len(runes) > messagePreviewMaxRunes {
			return string(runes[:messagePreviewMaxRunes]) + "..."
		}

		return trimmed
	case domain.MessageTypeVoice:
		return fmt.Sprintf("%s sent you a voice note.", firstName)
	case domain.MessageTypeGif:
		return fmt.Sprintf("%s sent you a GIF.", firstName)
	default:
		return ""
	}
}

func (s *service) ApplyScore(ctx context.Context, tx *sql.Tx, msg domain.Message) (*domain.ScoreSnapshot, error) {
	var result *domain.ScoreSnapshot

	switch msg.Type {
	case domain.MessageTypeText:
		var snap scoredomain.ScoreSnapshot

		contrib := scoredomain.Contribution{
			Type:    scoredomain.ContribText,
			TextLen: utils.CountTextLen(utils.TypePtrToString(msg.TextBody)),
		}

		var err error
		if tx == nil {
			snap, err = s.scoreService.Apply(ctx, msg.ConversationID, msg.SenderID, contrib)
			if err != nil {
				return nil, commonlogger.LogError(s.logger, "apply score", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
			}
		} else {
			snap, err = s.scoreService.ApplyViaTx(ctx, tx, msg.ConversationID, msg.SenderID, contrib)
			if err != nil {
				return nil, commonlogger.LogError(s.logger, "apply score via tx", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
			}
		}

		result = mapper.MapScoreDomainSnapShotToConversationDomain(snap)

	case domain.MessageTypeVoice:
		if msg.MediaSeconds == nil {
			s.logger.Error("apply score voice msg missing media_seconds", zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
			return nil, fmt.Errorf("apply score voice msg missing media_seconds")
		}

		secs := int(math.Round(*msg.MediaSeconds))
		if secs < 0 {
			secs = 0
		}

		contrib := scoredomain.Contribution{
			Type:    scoredomain.ContribVoice,
			Seconds: secs,
		}

		var snap scoredomain.ScoreSnapshot

		var err error

		if tx == nil {
			snap, err = s.scoreService.Apply(ctx, msg.ConversationID, msg.SenderID, contrib)
			if err != nil {
				return nil, commonlogger.LogError(s.logger, "apply voice score", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
			}
		} else {
			snap, err = s.scoreService.ApplyViaTx(ctx, tx, msg.ConversationID, msg.SenderID, contrib)
			if err != nil {
				return nil, commonlogger.LogError(s.logger, "apply voice score via tx", err, zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID))
			}
		}

		result = mapper.MapScoreDomainSnapShotToConversationDomain(snap)

		// todo: implement calls scoring
	case domain.MessageTypeSystem:
		// no scoring
	default:
		s.logger.Error("unsupported message type", zap.String("userID", msg.SenderID), zap.String("conversationID", msg.ConversationID), zap.String("messageType", string(msg.Type)))
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	return result, nil
}

func (s *service) Unmatch(ctx context.Context, userID, conversationID string, reason string) error {
	// Check if user is participant in conversation
	isParticipant, err := s.conversationRepo.IsConversationParticipant(ctx, conversationID, userID)
	if err != nil {
		return commonlogger.LogError(s.logger, "check conversation participant", err, zap.String("userID", userID), zap.String("conversationID", conversationID))
	}

	if !isParticipant {
		return storage.ErrNotConversationParticipant
	}

	// Get conversation to find the other user
	conversation, err := s.conversationRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get conversation", err, zap.String("userID", userID), zap.String("conversationID", conversationID))
	}

	if conversation == nil {
		return storage.ErrNonExistentConversation
	}

	// Determine the other user
	var otherUserID string
	if conversation.UserA == userID {
		otherUserID = conversation.UserB
	} else {
		otherUserID = conversation.UserA
	}

	// Use transaction to update match status and archive conversation
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin transaction", err)
	}

	defer func() { _ = tx.Rollback() }()

	// Update match status to unmatched with reason
	err = s.conversationRepo.SetMatchStatusWithReason(ctx, tx.Raw(), userID, otherUserID, string(domain.MatchStatusUnmatched), &reason)
	if err != nil {
		return commonlogger.LogError(s.logger, "set match status with reason", err, zap.String("userID", userID), zap.String("otherUserID", otherUserID))
	}

	// Archive the conversation
	_, err = s.conversationRepo.ArchiveConversationBetween(ctx, tx.Raw(), userID, otherUserID)
	if err != nil {
		return commonlogger.LogError(s.logger, "archive conversation", err, zap.String("userID", userID), zap.String("otherUserID", otherUserID))
	}

	err = tx.Commit()
	if err != nil {
		return commonlogger.LogError(s.logger, "commit transaction", err)
	}

	// Broadcast unmatch event to both users
	s.broadcastUnmatch(ctx, conversationID, userID)

	return nil
}

func (s *service) broadcastUnmatch(ctx context.Context, conversationID, actorID string) {
	// Get participants
	participants, err := s.conversationRepo.GetConversationParticipants(ctx, conversationID)
	if err != nil {
		s.logger.Error("broadcast unmatch: get participants", zap.Error(err))
		return
	}

	// Send event to both users
	evt := dto.Event{
		ID:        realtime.NewEventID(),
		Type:      "conversation.unmatched",
		ActorID:   actorID,
		Ts:        time.Now(),
		ContextID: conversationID,
		Data:      nil,
		Version:   1,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("broadcast unmatch: marshal event", zap.Error(err))
		return
	}

	for _, participant := range participants {
		s.hub.BroadcastToUser(participant.UserID, b)
	}
}
