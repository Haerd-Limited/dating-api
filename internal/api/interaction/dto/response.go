package dto

import (
	"errors"

	profiledto "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

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
	FreeToMatch        []Like `json:"free_to_match"`
	SlotsFull          []Like `json:"slots_full"`
	ActiveMatchesCount int64  `json:"active_matches_count"`
	MatchSlotLimit     int64  `json:"match_slot_limit"`
}

type Like struct {
	Profile            profiledto.ProfileCard `json:"profile"`
	Message            *Message               `json:"message"`
	Prompt             *Prompt                `json:"prompt"`
	TargetAtMatchLimit bool                   `json:"target_at_match_limit"`
	IsFavourited       bool                   `json:"is_favourited"`
}

type AddFavouriteRequest struct {
	WatchedUserID string `json:"watched_user_id"`
}

func (r AddFavouriteRequest) Validate() error {
	if r.WatchedUserID == "" {
		return errors.New("watched_user_id is required")
	}

	return nil
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
	MessageText  *string  `json:"message_text"`
	MessageType  *string  `json:"message_type"`
	VoiceNoteURL *string  `json:"voice_note_url,omitempty"`
	MediaSeconds *float64 `json:"media_seconds,omitempty"`
}
