package dto

import "time"

type GetConversationMessagesResponse struct {
	Messages []Message `json:"messages"`
}

type Conversation struct {
	ID string `json:"id"`
	// Match the user/person you matched with
	Match          Match     `json:"match"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	LastMessage    *Message  `json:"last_message"`
	UnreadCount    int       `json:"unread_count"`
}

type Match struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Emoji       string `json:"emoji"`
}

type Message struct {
	ID             int64     `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Type           string    `json:"type"`
	TextBody       *string   `json:"text_body"`
	MediaKey       *string   `json:"media_key"`
	MediaSeconds   *float64  `json:"media_seconds"`
	CreatedAt      time.Time `json:"created_at"`
	ClientMsgID    string    `json:"client_msg_id"`
}
