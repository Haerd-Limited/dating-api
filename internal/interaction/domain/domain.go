package domain

import "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"

type Swipe struct {
	TargetUserID   string
	Action         string
	PromptID       int64
	UserID         string
	Message        *string
	MessageType    *string
	VoiceNoteURL   *string
	IdempotencyKey *string
}

type Like struct {
	Profile profilecard.ProfileCard
	Message *Message
	Prompt  *Prompt
}
type Prompt struct {
	PromptID      int64
	Prompt        string
	VoiceNoteURL  string
	CoverPhotoUrl string
}

type Message struct {
	MessageText *string
	MessageType *string
}
