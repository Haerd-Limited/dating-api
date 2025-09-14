package conversation

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	"github.com/Haerd-Limited/dating-api/internal/profile"
)

type Service interface {
	GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error)
}

type service struct {
	logger             *zap.Logger
	conversationRepo   storage.ConversationRepository
	interactionService interaction.Service
	profileService     profile.Service
}

func NewConversationService(
	logger *zap.Logger,
	conversationRepo storage.ConversationRepository,
	interactionService interaction.Service,
	profileService profile.Service,
) Service {
	return &service{
		logger:             logger,
		conversationRepo:   conversationRepo,
		interactionService: interactionService,
		profileService:     profileService,
	}
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	// 1. Get Matches.
	// this is to make sure user's are matched since a user can't have a convo before matching. and if unmatched, convo should be gone
	matches, err := s.interactionService.GetMatches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches userID=%s: %w", userID, err)
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

		conversation, convoErr := s.GetConversation(ctx, userID, matchUserID)
		if convoErr != nil {
			return nil, fmt.Errorf("failed to get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
		}

		if conversation == nil {
			// create convo
			_, createConvoErr := s.conversationRepo.CreateConversation(ctx, userID, matchUserID)
			if createConvoErr != nil {
				return nil, fmt.Errorf("failed to create conversation userID=%s matchUserID=%s: %w", userID, matchUserID, createConvoErr)
			}

			conversation, convoErr = s.GetConversation(ctx, userID, matchUserID)
			if convoErr != nil {
				return nil, fmt.Errorf("failed to get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
			}
		}

		conversations = append(conversations, *conversation)
	}
	// todo: implement score/points system

	return conversations, nil
}

func (s *service) GetConversation(ctx context.Context, userID, matchID string) (*domain.Conversation, error) {
	conversationEntity, err := s.conversationRepo.GetConversationByUserIDs(ctx, userID, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation userID=%s matchID=%s: %w", userID, matchID, err)
	}

	if conversationEntity == nil {
		return nil, nil
	}

	// todo: might be overkill. might be better to just make a get profile/displayname repo method
	matchProfile, err := s.profileService.GetProfileCard(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match profile userID=%s matchID=%s: %w", userID, matchID, err)
	}

	var lastMessage *domain.Message

	if conversationEntity.LastMessageID.Valid {
		var lmErr error

		lastMessage, lmErr = s.getLastMessageByID(ctx, userID, matchID, conversationEntity.LastMessageID.Int64)
		if lmErr != nil {
			return nil, fmt.Errorf("failed to get last message userID=%s matchID=%s: %w", userID, matchID, err)
		}
	}

	return &domain.Conversation{
		ID: conversationEntity.ID,
		Match: domain.Match{
			ID:          matchID,
			DisplayName: matchProfile.DisplayName,
			Emoji:       "😊", // todo: allow users to set emoji at register and update profile
		},
		CreatedAt:      conversationEntity.CreatedAt,
		LastActivityAt: conversationEntity.LastActivityAt,
		// todo: populate last message or default to new convo if no messages.
		LastMessage: lastMessage,
		UnreadCount: 0, // todo: figure out what this means
	}, nil
}

func (s *service) getLastMessageByID(ctx context.Context, userID string, matchID string, lastMessageID int64) (*domain.Message, error) {
	lastMessageEntity, err := s.conversationRepo.GetLastMessageByID(ctx, lastMessageID)
	if err != nil {
		return &domain.Message{}, fmt.Errorf("failed to get last message userID=%s matchID=%s: %w", userID, matchID, err)
	}

	if lastMessageEntity == nil {
		return &domain.Message{}, nil
	}

	var messageType domain.MessageType

	switch lastMessageEntity.Type {
	case "text":
		messageType = domain.MessageTypeText
	case "voice":
		messageType = domain.MessageTypeVoice
	case "system":
		messageType = domain.MessageTypeSystem
	default:
		return &domain.Message{}, fmt.Errorf("unknown message type: %s", lastMessageEntity.Type)
	}

	var textBody *string
	if lastMessageEntity.TextBody.Valid {
		textBody = &lastMessageEntity.TextBody.String
	}

	var mediaKey *string
	if lastMessageEntity.MediaKey.Valid {
		mediaKey = &lastMessageEntity.MediaKey.String
	}

	mediaSeconds, msErr := strconv.ParseFloat(lastMessageEntity.MediaSeconds.String(), 64)
	if msErr != nil {
		return &domain.Message{}, fmt.Errorf("failed to parse media seconds: %w", msErr)
	}

	return &domain.Message{
		ID:             lastMessageEntity.ID,
		ConversationID: lastMessageEntity.ConversationID,
		SenderID:       lastMessageEntity.SenderID,
		Type:           messageType,
		TextBody:       textBody,
		MediaKey:       mediaKey,
		MediaSeconds:   &mediaSeconds,
		CreatedAt:      lastMessageEntity.CreatedAt,
	}, nil
}
