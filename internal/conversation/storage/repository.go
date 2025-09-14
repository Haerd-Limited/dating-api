package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/friendsofgo/errors"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type ConversationRepository interface {
	GetConversationByUserIDs(ctx context.Context, userA, userB string) (*entity.Conversation, error)
	GetLastMessageByID(ctx context.Context, lastMessageID int64) (*entity.Message, error)
	CreateConversation(ctx context.Context, userA, userB string) (*entity.Conversation, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
}

type repository struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) ConversationRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetConversationByUserIDs(ctx context.Context, userA, userB string) (*entity.Conversation, error) {
	result, err := entity.Conversations(
		qm.Where("user_a = ? AND user_b = ?", userA, userB),
		qm.Or2(qm.Where("user_a = ? AND user_b = ?", userB, userA)),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (r *repository) GetLastMessageByID(ctx context.Context, lastMessageID int64) (*entity.Message, error) {
	result, err := entity.Messages(
		entity.MessageWhere.ID.EQ(lastMessageID),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

// CreateConversation inserts a new conversation between two users.
// If a conversation already exists (userA↔userB), it returns that instead.
func (r *repository) CreateConversation(ctx context.Context, userA, userB string) (*entity.Conversation, error) {
	// Guard: no self-conversation
	if userA == userB {
		return nil, fmt.Errorf("cannot create conversation with self: %s", userA)
	}

	// Check if it already exists (either order)
	convo, err := entity.Conversations(
		entity.ConversationWhere.UserA.EQ(userA),
		entity.ConversationWhere.UserB.EQ(userB),
	).One(ctx, r.db)
	if err == nil {
		return convo, nil // already exists
	}

	convo, err = entity.Conversations(
		entity.ConversationWhere.UserA.EQ(userB),
		entity.ConversationWhere.UserB.EQ(userA),
	).One(ctx, r.db)
	if err == nil {
		return convo, nil // already exists (reverse order)
	}

	// Create new conversation
	now := time.Now()
	c := &entity.Conversation{
		UserA:          userA,
		UserB:          userB,
		CreatedAt:      now,
		LastActivityAt: now,
	}

	err = c.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("insert conversation failed: %w", err)
	}

	return c, nil
}

func (is *repository) GetMatches(ctx context.Context, userID string) ([]*entity.Match, error) {
	matches, err := entity.Matches(
		entity.MatchWhere.UserB.EQ(userID),
		qm.Or2(
			entity.MatchWhere.UserA.EQ(userID),
		),
	).All(ctx, is.db)
	if err != nil {
		return nil, err
	}

	return matches, nil
}
