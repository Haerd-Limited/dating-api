package realtime

import "context"

type ConversationAuth interface {
	IsConversationParticipant(ctx context.Context, conversationID, userID string) (bool, error)
}
