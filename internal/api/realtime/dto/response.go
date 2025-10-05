package dto

// From server → client
type ServerMsg struct {
	Type           string      `json:"type"` // "message.new" | "conversation.typing" | "pong" | ...
	ConversationID string      `json:"conversation_id,omitempty"`
	Payload        interface{} `json:"payload,omitempty"` // e.g., a MessageDTO
}
