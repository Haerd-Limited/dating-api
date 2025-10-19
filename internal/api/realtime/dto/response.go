package dto

import (
	"time"
)

// From server → client
type ServerMsg struct {
	Type           string      `json:"type"` // "message.new" | "conversation.typing" | "pong" | ...
	ConversationID string      `json:"conversation_id,omitempty"`
	Payload        interface{} `json:"payload,omitempty"` // e.g., a MessageDTO
}

type Event struct {
	// ID ULID() or snowflake for ordering
	ID string `json:"id"`
	// e.g. "like.created"
	Type string `json:"type"`
	// who caused it
	ActorID string    `json:"actorId"`
	Ts      time.Time `json:"ts"`
	// e.g. conversationID (when relevant)
	ContextID string `json:"contextId"`
	// type-specific payload
	Data any `json:"data"`
	// schema versioning
	Version int `json:"version"`
}
