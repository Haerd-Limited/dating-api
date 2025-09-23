package conversation

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/mapper"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
)

type Service interface {
	GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error)
	CreateConversation(ctx context.Context, userID, matchUserID string) error
	SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error)
}

type service struct {
	logger           *zap.Logger
	conversationRepo storage.ConversationRepository
	profileService   profile.Service
	flake            interface{ Next() int64 }
}

func NewConversationService(
	logger *zap.Logger,
	conversationRepo storage.ConversationRepository,
	profileService profile.Service,
	flake interface{ Next() int64 },
) Service {
	return &service{
		logger:           logger,
		conversationRepo: conversationRepo,
		profileService:   profileService,
		flake:            flake,
	}
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	// 1. Get Matches.
	// this is to make sure user's are matched since a user can't have a convo before matching. and if unmatched, convo should be gone
	matches, err := s.getMatches(ctx, userID)
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

		conversation, convoErr := s.GetConversationByUserIds(ctx, userID, matchUserID)
		if convoErr != nil {
			return nil, fmt.Errorf("failed to get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
		}

		if conversation == nil {
			// create convo
			_, createConvoErr := s.conversationRepo.CreateConversation(ctx, userID, matchUserID)
			if createConvoErr != nil {
				return nil, fmt.Errorf("failed to create conversation userID=%s matchUserID=%s: %w", userID, matchUserID, createConvoErr)
			}

			conversation, convoErr = s.GetConversationByUserIds(ctx, userID, matchUserID)
			if convoErr != nil {
				return nil, fmt.Errorf("failed to get conversation userID=%s matchUserID=%s: %w", userID, matchUserID, convoErr)
			}
		}

		conversations = append(conversations, *conversation)
	}
	// todo: implement score/points system

	return conversations, nil
}

func (s *service) GetConversationByUserIds(ctx context.Context, userID, matchID string) (*domain.Conversation, error) {
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
			return nil, fmt.Errorf("failed to get last message userID=%s matchID=%s: %w", userID, matchID, lmErr)
		}
	} else {
		systemMsg := fmt.Sprintf("Start the chat with %s", matchProfile.DisplayName)
		lastMessage = &domain.Message{
			ConversationID: conversationEntity.ID,
			Type:           domain.MessageTypeSystem,
			TextBody:       &systemMsg,
		}
	}

	return &domain.Conversation{
		ID: conversationEntity.ID,
		MatchedUser: domain.MatchedUser{
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

func (s *service) CreateConversation(ctx context.Context, userID, matchUserID string) error {
	_, err := s.conversationRepo.CreateConversation(ctx, userID, matchUserID)
	if err != nil {
		return fmt.Errorf("failed to create conversation userID=%s matchUserID=%s: %w", userID, matchUserID, err)
	}

	return nil
}

func (s *service) getLastMessageByID(ctx context.Context, userID string, matchID string, lastMessageID int64) (*domain.Message, error) {
	lastMessageEntity, err := s.conversationRepo.GetLastMessageByID(ctx, lastMessageID)
	if err != nil {
		return &domain.Message{}, fmt.Errorf("failed to get last message entity userID=%s matchID=%s: %w", userID, matchID, err)
	}

	if lastMessageEntity == nil {
		return nil, nil
	}

	msg, err := mapper.MapMessageEntityToDomain(*lastMessageEntity)
	if err != nil {
		return &domain.Message{}, fmt.Errorf("failed to map message entity userID=%s matchID=%s: %w", userID, matchID, err)
	}

	return &msg, nil
}

func (s *service) getMatches(ctx context.Context, userID string) ([]domain.Match, error) {
	matchEntities, err := s.conversationRepo.GetMatches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches userID=%s: %w", userID, err)
	}

	if len(matchEntities) == 0 {
		return []domain.Match{}, nil
	}

	return mapper.MapMatchEntitiesToDomain(matchEntities), nil
}

func (s *service) SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error) {
	msg.ID = s.flake.Next()
	msgEntity, err := s.conversationRepo.SendMessageViaTx(ctx, mapper.MapMessageDomainToEntity(msg))

	if err != nil {
		return domain.Message{}, fmt.Errorf("failed to send message userID=%s: %w", msg.SenderID, err)
	}

	result, err := mapper.MapMessageEntityToDomain(*msgEntity)
	if err != nil {
		return domain.Message{}, fmt.Errorf("failed to map message entity userID=%s: %w", msg.SenderID, err)
	}

	return result, nil
}
