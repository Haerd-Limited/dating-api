package mapper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func MapSwipeToDomain(swipe *entity.Swipe) domain.Swipe {
	result := domain.Swipe{
		ID:        swipe.ID,
		ActorID:   swipe.ActorID,
		TargetID:  swipe.TargetID,
		Action:    swipe.Action,
		CreatedAt: swipe.CreatedAt,
		UpdatedAt: swipe.UpdatedAt,
	}
	if swipe.PromptID.Valid {
		result.PromptID = swipe.PromptID.Ptr()
	}

	if swipe.Message.Valid {
		result.Message = swipe.Message.Ptr()
	}

	if swipe.VoicenoteURL.Valid {
		result.VoicenoteURL = swipe.VoicenoteURL.Ptr()
	}

	if swipe.IdempotencyKey.Valid {
		result.IdempotencyKey = swipe.IdempotencyKey.Ptr()
	}

	if swipe.MessageType.Valid {
		result.MessageType = swipe.MessageType.Ptr()
	}

	return result
}

func MapMatchEntitiesToDomain(matchEntities []*entity.Match) []domain.Match {
	var matches []domain.Match
	for _, matchEntity := range matchEntities {
		matches = append(matches, MapMatchEntityToDomain(matchEntity))
	}

	return matches
}

func MapMatchEntityToDomain(matchEntity *entity.Match) domain.Match {
	var revealedAt time.Time
	if matchEntity.RevealedAt.Valid {
		revealedAt = matchEntity.RevealedAt.Time
	}

	return domain.Match{
		ID:         matchEntity.ID,
		UserA:      matchEntity.UserA,
		UserB:      matchEntity.UserB,
		CreatedAt:  matchEntity.CreatedAt,
		RevealedAt: revealedAt,
	}
}

func MapMessageEntityToDomain(msg entity.Message) (domain.Message, error) {
	var textBody *string
	if msg.TextBody.Valid {
		textBody = &msg.TextBody.String
	}

	var mediaKey *string
	if msg.MediaKey.Valid {
		mediaKey = &msg.MediaKey.String
	}

	var mediaSeconds float64

	var msErr error
	if !msg.MediaSeconds.IsZero() {
		mediaSeconds, msErr = strconv.ParseFloat(msg.MediaSeconds.String(), 64)
		if msErr != nil {
			return domain.Message{}, fmt.Errorf("failed to parse media seconds: %w", msErr)
		}
	}

	var clientMSGID string
	if msg.ClientMSGID.Valid {
		clientMSGID = msg.ClientMSGID.String
	}

	return domain.Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           domain.MessageType(msg.Type),
		TextBody:       textBody,
		MediaUrl:       mediaKey,
		MediaSeconds:   &mediaSeconds,
		CreatedAt:      msg.CreatedAt,
		ClientMsgID:    clientMSGID,
	}, nil
}
