package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/friendsofgo/errors"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type ConversationRepository interface {
	GetConversationByUserIDs(ctx context.Context, userA, userB string) (*entity.Conversation, error)
	GetLastMessageByID(ctx context.Context, lastMessageID int64) (*entity.Message, error)
	CreateConversation(ctx context.Context, userA, userB string) (*entity.Conversation, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
	SendMessageViaTx(ctx context.Context, msg entity.Message) (*entity.Message, error)
	GetConversationByID(ctx context.Context, conversationID string) (*entity.Conversation, error)
}

type repository struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) ConversationRepository {
	return &repository{
		db: db,
	}
}

var (
	ErrNonExistentConversation    = errors.New("conversation does not exist")
	ErrNonExistentMatch           = errors.New("match does not exist")
	ErrMatchNotActive             = errors.New("match is not active")
	ErrNotConversationParticipant = errors.New("user is not a participant in the conversation")
)

func (r *repository) SendMessageViaTx(ctx context.Context, msg entity.Message) (*entity.Message, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// 1. check if conversation exists.
	convo, err := entity.Conversations(
		entity.ConversationWhere.ID.EQ(msg.ConversationID),
		qm.For("UPDATE"),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNonExistentConversation
		}

		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if msg.SenderID != convo.UserA && msg.SenderID != convo.UserB {
		return nil, ErrNotConversationParticipant // not a participant
	}
	// 2. if conversation exists, get and then check matches to see if status is active. if not, return forbidden
	match, err := entity.Matches(
		qm.Where("user_a = ? AND user_b = ? OR user_a = ? AND user_b = ?",
			convo.UserA, convo.UserB, convo.UserB, convo.UserA),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNonExistentMatch
		}

		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if match.Status != entity.MatchStatusActive {
		return nil, fmt.Errorf("%w : status=%s", ErrMatchNotActive, match.Status)
	}
	// 3. Insert message in message table and get messageID,
	err = msg.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}
	// 4. Set messageID as conversations messageID
	convo.LastMessageID = null.Int64From(msg.ID)
	convo.LastActivityAt = time.Now()

	_, err = convo.Update(ctx, tx, boil.Whitelist(
		entity.ConversationColumns.LastMessageID,
		entity.ConversationColumns.LastActivityAt,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return &msg, nil
}

func (r *repository) GetConversationByID(ctx context.Context, conversationID string) (*entity.Conversation, error) {
	convo, err := entity.Conversations(
		entity.ConversationWhere.ID.EQ(conversationID),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return convo, nil
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

func (r *repository) GetMatches(ctx context.Context, userID string) ([]*entity.Match, error) {
	matches, err := entity.Matches(
		entity.MatchWhere.UserB.EQ(userID),
		qm.Or2(
			entity.MatchWhere.UserA.EQ(userID),
		),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return matches, nil
}
