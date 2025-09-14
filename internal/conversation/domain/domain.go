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
	// Match the user/person you matched with
	Match          Match
	CreatedAt      time.Time
	LastActivityAt time.Time
	LastMessage    *Message
	UnreadCount    int
}

type Match struct {
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
}
