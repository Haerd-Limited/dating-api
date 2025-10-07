package dto

import "time"

type GetConversationMessagesResponse struct {
	Messages []Message `json:"messages"`
}

type SendMessageResponse struct {
	Messages Message `json:"message"`
}

type GetConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

type GetConversationScoreResponse struct {
	Score int `json:"score"`
}

type Conversation struct {
	ID string `json:"id"`
	// Match the user/person you matched with
	Match          Match         `json:"match"`
	CreatedAt      time.Time     `json:"created_at"`
	LastActivityAt time.Time     `json:"last_activity_at"`
	LastMessage    *Message      `json:"last_message"`
	UnreadCount    int           `json:"unread_count"`
	Score          ScoreSnapshot `json:"score_snapshot"`
}

type Match struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Emoji       string `json:"emoji"`
}

type Message struct {
	ID                     int64          `json:"id"`
	ConversationID         string         `json:"conversation_id"`
	SenderID               string         `json:"sender_id"`
	Type                   string         `json:"type"`
	TextBody               *string        `json:"text_body"`
	MediaKey               *string        `json:"media_key"`
	MediaSeconds           *float64       `json:"media_seconds"`
	CreatedAt              time.Time      `json:"created_at"`
	ClientMsgID            string         `json:"client_msg_id"`
	IsFirstMessage         bool           `json:"is_first_message"`
	LikedPrompt            *VoicePrompt   `json:"liked_prompt"` // populated if IsFirstMessage is true
	ResultingScoreSnapShot *ScoreSnapshot `json:"resulting_score_snapshot"`
}

type ScoreSnapshot struct {
	Threshold int  `json:"threshold"`
	Me        int  `json:"me"`
	Them      int  `json:"them"`
	Revealed  bool `json:"revealed"`
	Shared    int  `json:"shared"`
	CanReveal bool `json:"can_reveal"`
}

type VoicePrompt struct {
	ID            int64  `json:"id"`
	Prompt        string `json:"prompt"`
	CoverPhotoURL string `json:"cover_photo_url"`
	VoiceNoteURL  string `json:"voice_note_url"`
}
