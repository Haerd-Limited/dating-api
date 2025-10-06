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
	CreateConversation(ctx context.Context, userA, userB string, tx *sql.Tx) (*entity.Conversation, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
	SendMessage(ctx context.Context, msg entity.Message) (*entity.Message, error)
	SendMessageViaTx(ctx context.Context, tx *sql.Tx, msg entity.Message) (*entity.Message, error)
	GetConversationByID(ctx context.Context, conversationID string) (*entity.Conversation, error)
	GetMessagesByConversationID(ctx context.Context, conversationID, userID string) ([]*entity.Message, error)
	IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error)
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

func (r *repository) IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	isParticipant, err := entity.Conversations(
		entity.ConversationWhere.ID.EQ(conversationID),
		qm.Where("user_a = ? OR user_b = ?", userID, userID),
	).Exists(ctx, r.db)
	if err != nil {
		return false, err
	}

	return isParticipant, nil
}

func (r *repository) GetMessagesByConversationID(ctx context.Context, conversationID, userID string) ([]*entity.Message, error) {
	isParticipant, err := r.IsConversationParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("is conversation participant: %w", err)
	}

	if !isParticipant {
		return nil, ErrNotConversationParticipant
	}

	messages, err := entity.Messages(
		entity.MessageWhere.ConversationID.EQ(conversationID),
		qm.OrderBy(entity.MessageColumns.CreatedAt+" ASC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return messages, nil
}

// Standalone version that manages its own tx for non-aggregate use
func (r *repository) SendMessage(ctx context.Context, msg entity.Message) (*entity.Message, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	out, err := r.SendMessageViaTx(ctx, tx.Tx, msg)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return out, nil
}

// True “ViaTx” – uses caller’s tx
func (r *repository) SendMessageViaTx(ctx context.Context, tx *sql.Tx, msg entity.Message) (*entity.Message, error) {
	// 1) Lock conversation
	convo, err := entity.Conversations(
		entity.ConversationWhere.ID.EQ(msg.ConversationID),
		qm.For("UPDATE"),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNonExistentConversation
		}

		return nil, fmt.Errorf("get conversation: %w", err)
	}

	// 2) Participant guard
	if msg.SenderID != convo.UserA && msg.SenderID != convo.UserB {
		return nil, ErrNotConversationParticipant
	}

	// 3) Match status (optional FOR UPDATE if you need to lock it)
	match, err := entity.Matches(
		qm.Where("(user_a = ? AND user_b = ?) OR (user_a = ? AND user_b = ?)",
			convo.UserA, convo.UserB, convo.UserB, convo.UserA),
		qm.For("UPDATE"),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNonExistentMatch
		}

		return nil, fmt.Errorf("get match: %w", err)
	}

	if match.Status != entity.MatchStatusActive {
		return nil, fmt.Errorf("%w: status=%s", ErrMatchNotActive, match.Status)
	}
	// 3. Insert message in message table and get messageID,
	err = msg.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}

	// 5) Update conversation tail
	convo.LastMessageID = null.Int64From(msg.ID)
	convo.LastActivityAt = time.Now().UTC()

	_, err = convo.Update(ctx, tx, boil.Whitelist(
		entity.ConversationColumns.LastMessageID,
		entity.ConversationColumns.LastActivityAt,
	))
	if err != nil {
		return nil, fmt.Errorf("update conversation: %w", err)
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
func (r *repository) CreateConversation(ctx context.Context, userA, userB string, tx *sql.Tx) (*entity.Conversation, error) {
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

	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	err = c.Insert(ctx, exec, boil.Infer())
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
