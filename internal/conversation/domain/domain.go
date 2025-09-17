package domain

import "time"

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeVoice  MessageType = "voice"
	MessageTypeSystem MessageType = "system"
)

type Conversation struct {
	ID string
	// MatchedUser the user/person you matched with
	MatchedUser    MatchedUser
	CreatedAt      time.Time
	LastActivityAt time.Time
	LastMessage    *Message
	UnreadCount    int
}

type MatchedUser struct {
	ID          string
	DisplayName string
	Emoji       string
}

type Message struct {
	ID             int64
	ConversationID string
	SenderID       string
	Type           MessageType
	TextBody       *string
	MediaKey       *string
	MediaSeconds   *float64
	CreatedAt      time.Time
	ClientMsgID    string
}

type MatchStatus string

const (
	MatchStatusActive    MatchStatus = "active"
	MatchStatusUnmatched MatchStatus = "unmatched"
	MatchStatusBlocked   MatchStatus = "blocked"
)

type Match struct {
	ID         string
	UserA      string
	UserB      string
	CreatedAt  time.Time
	RevealedAt time.Time
}
