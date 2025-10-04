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
	Likes []Like `json:"likes"`
}

type Like struct {
	Profile dto.ProfileCard `json:"profile"`
	Message *Message        `json:"message"`
}

type Message struct {
	MessageText *string `json:"message_text"`
	MessageType *string `json:"message_type"`
}
