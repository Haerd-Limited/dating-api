package domain

import (
	"time"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type MessageType string

const (
	MessageTypeText   MessageType = constants.MessageTypeText
	MessageTypeVoice  MessageType = constants.MessageTypeVoice
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
	IsFirstMessage bool
	LikedPrompt    *VoicePrompt // populated if IsFirstMessage is true
}

type VoicePrompt struct {
	ID            int64
	Prompt        string
	CoverPhotoURL string
	VoiceNoteURL  string
}

type Swipe struct {
	ID             int64
	ActorID        string
	TargetID       string
	Action         string
	IdempotencyKey *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MessageType    *string
	Message        *string
	VoicenoteURL   *string
	PromptID       *int64
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
