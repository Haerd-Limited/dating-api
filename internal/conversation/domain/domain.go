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
	MessageTypeGif    MessageType = "gif"
)

type Conversation struct {
	ID string
	// MatchedUser the user/person you matched with
	MatchedUser    MatchedUser
	CreatedAt      time.Time
	LastActivityAt time.Time
	LastMessage    *Message
	UnreadCount    int
	DateMode       bool
	Photos         []Photo
}

type MatchedUser struct {
	ID          string
	DisplayName string
	Emoji       string
	Theme       Theme
}

type Theme struct {
	BaseHex string
	Palette []string
}

type Message struct {
	ID                     int64
	ConversationID         string
	SenderID               string
	Type                   MessageType
	TextBody               *string
	MediaUrl               *string
	MediaSeconds           *float64
	CreatedAt              time.Time
	ClientMsgID            string
	IsFirstMessage         bool
	LikedPrompt            *VoicePrompt // populated if IsFirstMessage is true
	ResultingScoreSnapShot *ScoreSnapshot
}

type ScoreSnapshot struct {
	Threshold int
	Me        int
	Them      int
	Revealed  bool
	Shared    int  // min(Me, Them)
	CanReveal bool // Shared >= Threshold
}

type VoicePrompt struct {
	ID                    int64
	Prompt                string
	CoverMediaURL         string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64
	VoiceNoteURL          string
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

type RevealStatus string

const (
	RevealStatusPending   RevealStatus = "pending"
	RevealStatusExpired   RevealStatus = "expired"
	RevealStatusConfirmed RevealStatus = "confirmed"
	RevealStatusCancelled RevealStatus = "cancelled"
)

type RevealRequest struct {
	ConversationID string
	InitiatorID    string
	RequestedAt    time.Time
	ExpiresAt      time.Time
	Status         RevealStatus
}

type RevealDecision string

const (
	RevealDecisionContinue RevealDecision = constants.RevealDecisionContinue
	RevealDecisionDate     RevealDecision = constants.RevealDecisionDate
	RevealDecisionUnmatch  RevealDecision = constants.RevealDecisionUnmatch
)

type Photo struct {
	URL       string
	IsPrimary bool
	Position  int16
}
