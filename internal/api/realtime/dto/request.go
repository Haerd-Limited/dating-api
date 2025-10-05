package dto

// From client → server
type ClientMsg struct {
	Type           string `json:"type"` // "subscribe.conversation" | "unsubscribe.conversation" | "ping"
	ConversationID string `json:"conversation_id,omitempty"`
}
