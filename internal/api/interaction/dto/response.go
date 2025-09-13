package dto

type Match struct {
	UserID         string `json:"user_id"`
	DisplayName    string `json:"display_name"`
	MessagePreview string `json:"message_preview"`
	Emoji          string `json:"emoji"`
	Reveal         bool   `json:"reveal"`
	RevealProgress int    `json:"reveal_progress"`
}
