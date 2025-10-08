package conversation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/mapper"
	"github.com/Haerd-Limited/dating-api/internal/conversation/score"
	scoredomain "github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	storage3 "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

type Service interface {
	GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error)
	GetConversationScore(ctx context.Context, userID string, convoID string) (int, error)
	CreateConversation(ctx context.Context, userID, matchUserID string) error
	CreateConversationViaTx(ctx context.Context, userID, matchUserID string, tx *sql.Tx) (string, error)
	SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error)
	SendMessageViaTx(ctx context.Context, tx *sql.Tx, msg domain.Message) (domain.Message, error)
	GetMessages(ctx context.Context, convoID string, userID string) ([]domain.Message, error)
	IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error)
}

type service struct {
	logger           *zap.Logger
	conversationRepo storage.ConversationRepository
	profileService   profile.Service
	flake            interface{ Next() int64 }
	hub              realtime.Broadcaster
	interactionRepo  storage3.InteractionRepository
	scoreService     score.Service
}

func NewConversationService(
	logger *zap.Logger,
	conversationRepo storage.ConversationRepository,
	profileService profile.Service,
	flake interface{ Next() int64 },
	hub realtime.Broadcaster,
	interactionRepo storage3.InteractionRepository,
	scoreService score.Service,
) Service {
	return &service{
		logger:           logger,
		conversationRepo: conversationRepo,
		profileService:   profileService,
		flake:            flake,
		hub:              hub,
		interactionRepo:  interactionRepo,
		scoreService:     scoreService,
	}
}

var ErrFirstMessageMissingPromptID = errors.New("first message missing prompt id")

func (s *service) GetConversationScore(ctx context.Context, userID string, convoID string) (int, error) {
	_, _, shared, err := s.scoreService.GetScores(ctx, convoID, userID)
	if err != nil {
		return 0, fmt.Errorf("get scores userID= %s convoID= %s: %w", userID, convoID, err)
	}

	return shared, nil
}

func (s *service) IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	return s.conversationRepo.IsConversationParticipant(ctx, conversationID, userID)
}

func (s *service) GetMessages(ctx context.Context, convoID string, userID string) ([]domain.Message, error) {
	messageEntities, err := s.conversationRepo.GetMessagesByConversationID(ctx, convoID, userID)
	if err != nil {
		return nil, fmt.Errorf("get messages userID=%s convoID=%s: %w", userID, convoID, err)
	}

	// Assuming that both participants simply liked each-other.
	if len(messageEntities) == 0 {
		var firstLike domain.Swipe

		firstLike, err = s.getFirstLikeByConvoID(ctx, convoID)
		if err != nil {
			return nil, fmt.Errorf("get first like by convo id userID=%s convoID=%s: %w", userID, convoID, err)
		}

		_, err = s.SendMessage(ctx, domain.Message{
			ConversationID: convoID,
			SenderID:       firstLike.ActorID,
			Type:           domain.MessageType(*firstLike.MessageType),
			TextBody:       firstLike.Message,
			ClientMsgID:    uuid.New().String(),
		})
		if err != nil {
			return nil, fmt.Errorf("send message userID=%s convoID=%s: %w", userID, convoID, err)
		}

		messageEntities, err = s.conversationRepo.GetMessagesByConversationID(ctx, convoID, userID)
		if err != nil {
			return nil, fmt.Errorf("get messages userID=%s convoID=%s: %w", userID, convoID, err)
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
			return nil, fmt.Errorf("map message entity userID=%s convoID=%s: %w", userID, convoID, err)
		}

		// get prompt for first message
		if index == firstSentMessageIndex {
			msg.LikedPrompt, err = s.getLikedVoicePromptByConvoID(ctx, convoID, userID)
			if err != nil {
				return nil, fmt.Errorf("get liked prompt by convo id userID=%s convoID=%s: %w", userID, convoID, err)
			}

			msg.IsFirstMessage = true
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	// 1. Get Matches.
	// this is to make sure user's are matched since a user can't have a convo before matching. and if unmatched, convo should be gone
	matches, err := s.getMatches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get matches userID=%s: %w", userID, err)
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
			return nil, fmt.Errorf("get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
		}

		if conversation == nil {
			// create convo
			_, createConvoErr := s.conversationRepo.CreateConversation(ctx, userID, matchUserID, nil)
			if createConvoErr != nil {
				return nil, fmt.Errorf("create conversation userID=%s matchUserID=%s: %w", userID, matchUserID, createConvoErr)
			}

			conversation, convoErr = s.getConversationByUserIds(ctx, userID, matchUserID)
			if convoErr != nil {
				return nil, fmt.Errorf("get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
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
		return fmt.Errorf("create conversation userID=%s matchUserID=%s: %w", userID, matchUserID, err)
	}

	return nil
}

func (s *service) CreateConversationViaTx(ctx context.Context, userID, matchUserID string, tx *sql.Tx) (string, error) {
	convoEntity, err := s.conversationRepo.CreateConversation(ctx, userID, matchUserID, tx)
	if err != nil {
		return "", fmt.Errorf("create conversation via tx userID=%s matchUserID=%s: %w", userID, matchUserID, err)
	}

	return convoEntity.ID, nil
}

// For general API calls
func (s *service) SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error) {
	msg.ID = s.flake.Next()
	ent := mapper.MapMessageDomainToEntity(msg)

	e, err := s.conversationRepo.SendMessage(ctx, ent) // standalone version
	if err != nil {
		return domain.Message{}, fmt.Errorf("send message userID=%s conversationID=%s: %w", msg.SenderID, msg.ConversationID, err)
	}

	result, err := mapper.MapMessageEntityToDomain(*e)
	if err != nil {
		return domain.Message{}, fmt.Errorf("map message entity userID=%s: %w", msg.SenderID, err)
	}

	result.ResultingScoreSnapShot, err = s.ApplyScore(ctx, nil, result)
	if err != nil {
		return domain.Message{}, fmt.Errorf("apply score userID=%s convoID=%s: %w", msg.SenderID, msg.ConversationID, err)
	}

	s.sendMessageToConversation(result)
	// todo: broadcast score update

	return result, nil
}

// For aggregate flows (e.g., called from CreateSwipe)
func (s *service) SendMessageViaTx(ctx context.Context, tx *sql.Tx, msg domain.Message) (domain.Message, error) {
	msg.ID = s.flake.Next()
	ent := mapper.MapMessageDomainToEntity(msg)

	e, err := s.conversationRepo.SendMessageViaTx(ctx, tx, ent)
	if err != nil {
		return domain.Message{}, fmt.Errorf("send message via tx userID=%s conversationID=%s: %w", msg.SenderID, msg.ConversationID, err)
	}

	result, err := mapper.MapMessageEntityToDomain(*e)
	if err != nil {
		return domain.Message{}, fmt.Errorf("map message entity userID=%s: %w", msg.SenderID, err)
	}

	result.ResultingScoreSnapShot, err = s.ApplyScore(ctx, tx, result)
	if err != nil {
		return domain.Message{}, fmt.Errorf("apply score via tx userID=%s convoID=%s: %w", msg.SenderID, msg.ConversationID, err)
	}

	s.sendMessageToConversation(result)

	return result, nil
}

func (s *service) ApplyScore(ctx context.Context, tx *sql.Tx, msg domain.Message) (*domain.ScoreSnapshot, error) {
	var result *domain.ScoreSnapshot

	switch msg.Type {
	case domain.MessageTypeText:
		var snap scoredomain.ScoreSnapshot

		var err error
		if tx == nil {
			snap, err = s.scoreService.Apply(ctx, msg.ConversationID, msg.SenderID, scoredomain.Contribution{
				Type:    scoredomain.ContribText,
				TextLen: utils.CountTextLen(utils.TypePtrToString(msg.TextBody)),
			})
			if err != nil {
				return nil, fmt.Errorf("apply score userID=%s convoID=%s: %w", msg.SenderID, msg.ConversationID, err)
			}
		} else {
			snap, err = s.scoreService.ApplyViaTx(ctx, tx, msg.ConversationID, msg.SenderID, scoredomain.Contribution{
				Type:    scoredomain.ContribText,
				TextLen: utils.CountTextLen(utils.TypePtrToString(msg.TextBody)),
			})
			if err != nil {
				return nil, fmt.Errorf("apply score via tx userID=%s convoID=%s: %w", msg.SenderID, msg.ConversationID, err)
			}
		}

		result = mapper.MapScoreDomainSnapShotToConversationDomain(snap)

		// todo: implement voicenote scoring
		// todo: implement calls scoring
	case domain.MessageTypeSystem:
		// no scoring
	default:
		return nil, fmt.Errorf("unsupported message type userID=%s convoID=%s: %s",
			msg.SenderID, msg.ConversationID, msg.Type)
	}

	return result, nil
}
