package dto

import "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"

type Match struct {
	UserID         string `json:"user_id"`
	DisplayName    string `json:"display_name"`
	MessagePreview string `json:"message_preview"`
	Emoji          string `json:"emoji"`
	Reveal         bool   `json:"reveal"`
	RevealProgress int    `json:"reveal_progress"`
}

type SwipesResponse struct {
	Result string `json:"result"`
}

type GetLikesResponse struct {
	Verified   []Like `json:"verified"`
	Unverified []Like `json:"unverified"`
}

type Like struct {
	Profile dto.ProfileCard `json:"profile"`
	Message *Message        `json:"message"`
	Prompt  *Prompt         `json:"prompt"`
}

type Prompt struct {
	PromptID              int64    `json:"prompt_id"`
	Prompt                string   `json:"prompt"`
	VoiceNoteURL          string   `json:"voice_note_url"`
	CoverMediaURL         string   `json:"cover_media_url"`
	CoverMediaType        *string  `json:"cover_media_type,omitempty"`
	CoverMediaAspectRatio *float64 `json:"cover_media_aspect_ratio,omitempty"`
}

type Message struct {
	MessageText *string `json:"message_text"`
	MessageType *string `json:"message_type"`
}
