package domain

import "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"

type Swipe struct {
	TargetUserID   string
	Action         string
	UserID         string
	Message        *string
	MessageType    *string
	VoiceNoteURL   *string
	IdempotencyKey *string
}

type Like struct {
	Profile profilecard.ProfileCard
	Message *Message
}

type Message struct {
	MessageText *string
	MessageType *string
}
