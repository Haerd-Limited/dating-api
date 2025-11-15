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
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type ConversationRepository interface {
	GetConversationByUserIDs(ctx context.Context, userA, userB string) (*entity.Conversation, error)
	GetLastMessageByID(ctx context.Context, lastMessageID int64) (*entity.Message, error)
	CreateConversation(ctx context.Context, userA, userB string, tx *sql.Tx) (*entity.Conversation, error)
	GetMatches(ctx context.Context, userID string) ([]*entity.Match, error)
	SetMatchStatus(ctx context.Context, tx *sql.Tx, userA, userB, status string) error
	SendMessageViaTx(ctx context.Context, tx *sql.Tx, msg entity.Message) (*entity.Message, error)
	GetConversationByID(ctx context.Context, conversationID string) (*entity.Conversation, error)
	GetMessagesByConversationID(ctx context.Context, conversationID, userID string) ([]*entity.Message, error)
	IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error)
	UpdateUserConversationScore(ctx context.Context, tx *sql.Tx, conversationID, userID string, earned int) error
	GetScoreSettings(ctx context.Context) (entity.ScoringSetting, error)
	GetScoringText(ctx context.Context) (entity.ScoringText, error)
	GetScoringBonuses(ctx context.Context) (entity.ScoringBonuse, error)
	GetScoringCall(ctx context.Context) (entity.ScoringCall, error)
	GetScoringVoice(ctx context.Context) (entity.ScoringVoice, error)
	GetUserConversationScore(ctx context.Context, userID, convoID string, tx *sql.Tx) (int, error)
	GetOtherParticipantConversationScore(ctx context.Context, userID, convoID string, tx *sql.Tx) (int, error)
	SetConversationToRevealed(ctx context.Context, tx *sql.Tx, conversationID string) error
	CreateConversationScores(ctx context.Context, convoID, userID, matchedUserID string, tx *sql.Tx) error
	GetConversationParticipants(ctx context.Context, conversationID string) ([]*entity.ConversationParticipant, error)
	CreateRevealRequest(ctx context.Context, conversationID, userID string, expiresAt time.Time) error
	GetRevealRequest(ctx context.Context, conversationID string) (*entity.RevealRequest, error)
	UpdateRevealRequestStatus(ctx context.Context, conversationID string, status string) error
	SaveRevealDecision(ctx context.Context, conversationID, userID string, decision string) error
	GetRevealDecisions(ctx context.Context, conversationID string) ([]*entity.RevealDecision, error)
	SetDateMode(ctx context.Context, conversationID string, dateMode bool) error
	ArchiveConversationBetween(ctx context.Context, tx *sql.Tx, userA, userB string) (*string, error)
	MarkConversationMessagesAsRead(ctx context.Context, conversationID, userID string, tx *sql.Tx) error
	GetUnreadCount(ctx context.Context, conversationID, userID string) (int, error)
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
	ErrClientMsgIDNotUnique       = errors.New("client message ID is not unique")
	ErrRevealAlreadyInitiated     = errors.New("reveal request already exists for this conversation")
)

func (r *repository) GetConversationParticipants(ctx context.Context, conversationID string) ([]*entity.ConversationParticipant, error) {
	participants, err := entity.ConversationParticipants(
		entity.ConversationParticipantWhere.ConversationID.EQ(conversationID),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *repository) SetConversationToRevealed(ctx context.Context, tx *sql.Tx, conversationID string) error {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	convo, err := entity.Conversations(
		entity.ConversationWhere.ID.EQ(conversationID),
		entity.ConversationWhere.VisibilityState.EQ(constants.VisibilityStateHidden),
		qm.For("UPDATE"),
	).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("get conversation: %w", err)
	}

	convo.RevealAt = null.TimeFrom(time.Now().UTC())
	convo.VisibilityState = constants.VisibilityStateVisible

	_, err = convo.Update(ctx, exec, boil.Whitelist(
		entity.ConversationColumns.RevealAt,
		entity.ConversationColumns.VisibilityState,
	))
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	return nil
}

func (r *repository) GetOtherParticipantConversationScore(ctx context.Context, userID, convoID string, tx *sql.Tx) (int, error) {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	convoParticipants, err := entity.ConversationParticipants(
		entity.ConversationParticipantWhere.UserID.NEQ(userID),
		entity.ConversationParticipantWhere.ConversationID.EQ(convoID),
	).One(ctx, exec)
	if err != nil {
		return 0, err
	}

	return convoParticipants.Score, nil
}

func (r *repository) CreateConversationScores(ctx context.Context, convoID, userID, matchedUserID string, tx *sql.Tx) error {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	userScore := entity.ConversationParticipant{
		UserID:         userID,
		Score:          0,
		ScoreLifetime:  0,
		ConversationID: convoID,
	}

	err := userScore.Insert(ctx, exec, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert user score: %w", err)
	}

	matchedUserScore := entity.ConversationParticipant{
		UserID:         matchedUserID,
		Score:          0,
		ScoreLifetime:  0,
		ConversationID: convoID,
	}

	err = matchedUserScore.Insert(ctx, exec, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert matched user score: %w", err)
	}

	return nil
}

func (r *repository) GetUserConversationScore(ctx context.Context, userID, convoID string, tx *sql.Tx) (int, error) {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	convoParticipants, err := entity.ConversationParticipants(
		entity.ConversationParticipantWhere.UserID.EQ(userID),
		entity.ConversationParticipantWhere.ConversationID.EQ(convoID),
	).One(ctx, exec)
	if err != nil {
		return 0, err
	}

	return convoParticipants.Score, nil
}

func (r *repository) GetScoringVoice(ctx context.Context) (entity.ScoringVoice, error) {
	setting, err := entity.ScoringVoices(
		entity.ScoringVoiceWhere.ID.EQ(1),
	).One(ctx, r.db)
	if err != nil {
		return entity.ScoringVoice{}, err
	}

	return *setting, nil
}

func (r *repository) GetScoringCall(ctx context.Context) (entity.ScoringCall, error) {
	setting, err := entity.ScoringCalls(
		entity.ScoringCallWhere.ID.EQ(1),
	).One(ctx, r.db)
	if err != nil {
		return entity.ScoringCall{}, err
	}

	return *setting, nil
}

func (r *repository) GetScoringBonuses(ctx context.Context) (entity.ScoringBonuse, error) {
	setting, err := entity.ScoringBonuses(
		entity.ScoringBonuseWhere.ID.EQ(1),
	).One(ctx, r.db)
	if err != nil {
		return entity.ScoringBonuse{}, err
	}

	return *setting, nil
}

func (r *repository) GetScoringText(ctx context.Context) (entity.ScoringText, error) {
	setting, err := entity.ScoringTexts(
		entity.ScoringTextWhere.ID.EQ(1),
	).One(ctx, r.db)
	if err != nil {
		return entity.ScoringText{}, err
	}

	return *setting, nil
}

func (r *repository) GetScoreSettings(ctx context.Context) (entity.ScoringSetting, error) {
	setting, err := entity.ScoringSettings(
		entity.ScoringSettingWhere.ID.EQ(1),
	).One(ctx, r.db)
	if err != nil {
		return entity.ScoringSetting{}, err
	}

	return *setting, nil
}

func (r *repository) UpdateUserConversationScore(ctx context.Context, tx *sql.Tx, conversationID, userID string, earned int) error {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	convoParticipant, err := entity.ConversationParticipants(
		entity.ConversationParticipantWhere.ConversationID.EQ(conversationID),
		entity.ConversationParticipantWhere.UserID.EQ(userID),
		qm.For("UPDATE"),
	).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("get conversation participant: %w", err)
	}

	convoParticipant.Score += earned
	convoParticipant.ScoreLifetime += earned
	convoParticipant.LastContribAt = null.TimeFrom(time.Now().UTC())

	_, err = convoParticipant.Update(ctx, exec, boil.Whitelist(
		entity.ConversationParticipantColumns.Score,
		entity.ConversationParticipantColumns.ScoreLifetime,
		entity.ConversationParticipantColumns.LastContribAt,
	))
	if err != nil {
		return fmt.Errorf("update conversation participant: %w", err)
	}

	return nil
}

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
func (r *repository) sendMessage(ctx context.Context, msg entity.Message) (*entity.Message, error) {
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
	var result *entity.Message

	var err error
	if tx == nil {
		result, err = r.sendMessage(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}

		return result, nil
	}
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "ux_messages_sender_clientmsg":
				return nil, ErrClientMsgIDNotUnique
			default:
				return nil, fmt.Errorf("failed to insert message: %w", err)
			}
		}

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

	result = &msg

	return result, nil
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
	now := time.Now().UTC()
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

	err = r.CreateConversationScores(ctx, c.ID, userA, userB, tx)
	if err != nil {
		return nil, fmt.Errorf("create conversation scores failed: %w", err)
	}

	return c, nil
}

func (r *repository) GetMatches(ctx context.Context, userID string) ([]*entity.Match, error) {
	matches, err := entity.Matches(
		entity.MatchWhere.UserB.EQ(userID),
		qm.Or2(
			entity.MatchWhere.UserA.EQ(userID),
		),
		entity.MatchWhere.Status.EQ(string(entity.MatchStatusActive)),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (r *repository) SetMatchStatus(ctx context.Context, tx *sql.Tx, userA, userB, status string) error {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	match, err := entity.Matches(
		qm.Where("(user_a = ? AND user_b = ?) OR (user_a = ? AND user_b = ?)", userA, userB, userB, userA),
		qm.For("UPDATE"),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("get match for status update: %w", err)
	}

	match.Status = status

	_, err = match.Update(ctx, exec, boil.Whitelist(entity.MatchColumns.Status))
	if err != nil {
		return fmt.Errorf("update match status: %w", err)
	}

	return nil
}

func (r *repository) ArchiveConversationBetween(ctx context.Context, tx *sql.Tx, userA, userB string) (*string, error) {
	var exec boil.ContextExecutor
	if tx != nil {
		exec = tx
	} else {
		exec = r.db
	}

	convo, err := entity.Conversations(
		qm.Where("(user_a = ? AND user_b = ?) OR (user_a = ? AND user_b = ?)", userA, userB, userB, userA),
		qm.For("UPDATE"),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get conversation for archive: %w", err)
	}

	convo.VisibilityState = constants.VisibilityStateHidden
	convo.RevealAt = null.Time{}

	_, err = convo.Update(ctx, exec, boil.Whitelist(
		entity.ConversationColumns.VisibilityState,
		entity.ConversationColumns.RevealAt,
	))
	if err != nil {
		return nil, fmt.Errorf("update conversation visibility state: %w", err)
	}

	id := convo.ID

	return &id, nil
}

func (r *repository) CreateRevealRequest(ctx context.Context, conversationID, userID string, expiresAt time.Time) error {
	revealRequest := &entity.RevealRequest{
		ConversationID: conversationID,
		InitiatorID:    userID,
		RequestedAt:    time.Now().UTC(),
		ExpiresAt:      expiresAt,
		Status:         "pending",
	}

	err := revealRequest.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert reveal request failed: %w", err)
	}

	return nil
}

func (r *repository) GetRevealRequest(ctx context.Context, conversationID string) (*entity.RevealRequest, error) {
	revealRequest, err := entity.RevealRequests(
		entity.RevealRequestWhere.ConversationID.EQ(conversationID),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get reveal request failed: %w", err)
	}

	return revealRequest, nil
}

func (r *repository) UpdateRevealRequestStatus(ctx context.Context, conversationID string, status string) error {
	revealRequest, err := entity.RevealRequests(
		entity.RevealRequestWhere.ConversationID.EQ(conversationID),
	).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("get reveal request for update failed: %w", err)
	}

	revealRequest.Status = status

	_, err = revealRequest.Update(ctx, r.db, boil.Whitelist(entity.RevealRequestColumns.Status))
	if err != nil {
		return fmt.Errorf("update reveal request status failed: %w", err)
	}

	return nil
}

func (r *repository) SaveRevealDecision(ctx context.Context, conversationID, userID string, decision string) error {
	revealDecision := &entity.RevealDecision{
		ConversationID: conversationID,
		UserID:         userID,
		Decision:       decision,
		DecidedAt:      time.Now().UTC(),
	}

	err := revealDecision.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert reveal decision failed: %w", err)
	}

	return nil
}

func (r *repository) GetRevealDecisions(ctx context.Context, conversationID string) ([]*entity.RevealDecision, error) {
	decisions, err := entity.RevealDecisions(
		entity.RevealDecisionWhere.ConversationID.EQ(conversationID),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get reveal decisions failed: %w", err)
	}

	return decisions, nil
}

func (r *repository) SetDateMode(ctx context.Context, conversationID string, dateMode bool) error {
	// Get the match for this conversation
	match, err := entity.Matches(
		qm.Where("(user_a, user_b) IN (SELECT user_a, user_b FROM conversations WHERE id = ?)", conversationID),
		qm.Or2(qm.Where("(user_b, user_a) IN (SELECT user_a, user_b FROM conversations WHERE id = ?)", conversationID)),
	).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("get match for date mode update failed: %w", err)
	}

	match.DateMode = dateMode

	_, err = match.Update(ctx, r.db, boil.Whitelist(entity.MatchColumns.DateMode))
	if err != nil {
		return fmt.Errorf("update match date mode failed: %w", err)
	}

	return nil
}

func (r *repository) MarkConversationMessagesAsRead(ctx context.Context, conversationID, userID string, tx *sql.Tx) error {
	// Insert read receipts (status=2) for all messages in the conversation
	// where the message was sent by someone other than the user
	// and doesn't already have a read receipt for this user
	query := `
		INSERT INTO message_receipts (message_id, user_id, status, at)
		SELECT m.id, $2, 2, NOW()
		FROM messages m
		WHERE m.conversation_id = $1
		  AND m.sender_id != $2
		  AND m.type != 'system'
		  AND NOT EXISTS (
			SELECT 1 FROM message_receipts mr
			WHERE mr.message_id = m.id
			  AND mr.user_id = $2
			  AND mr.status = 2
		  )
		ON CONFLICT (message_id, user_id, status) DO NOTHING
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, conversationID, userID)
	} else {
		_, err = r.db.ExecContext(ctx, query, conversationID, userID)
	}

	if err != nil {
		return fmt.Errorf("mark conversation messages as read: %w", err)
	}

	return nil
}

func (r *repository) GetUnreadCount(ctx context.Context, conversationID, userID string) (int, error) {
	// Count messages where:
	// - sender_id != userID (messages sent by others)
	// - type != 'system' (exclude system messages)
	// - No read receipt (status=2) exists for this user
	query := `
		SELECT COUNT(DISTINCT m.id)
		FROM messages m
		WHERE m.conversation_id = $1
		  AND m.sender_id != $2
		  AND m.type != 'system'
		  AND NOT EXISTS (
			SELECT 1 FROM message_receipts mr
			WHERE mr.message_id = m.id
			  AND mr.user_id = $2
			  AND mr.status = 2
		  )
	`

	var count int

	err := r.db.GetContext(ctx, &count, query, conversationID, userID)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}

	return count, nil
}
