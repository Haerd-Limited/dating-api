package mapper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

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
		MediaKey:       mediaKey,
		MediaSeconds:   &mediaSeconds,
		CreatedAt:      msg.CreatedAt,
		ClientMsgID:    clientMSGID,
	}, nil
}
