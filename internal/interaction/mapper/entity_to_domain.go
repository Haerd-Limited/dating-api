package mapper

import (
	"fmt"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	interactiondomain "github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

func SwipeMessageToDomain(swipe *entity.Swipe) (*interactiondomain.Message, error) {
	if swipe == nil {
		return nil, nil
	}

	msg := &interactiondomain.Message{}

	if swipe.Message.Valid && swipe.MessageType.Valid {
		msg.MessageText = swipe.Message.Ptr()
		msg.MessageType = swipe.MessageType.Ptr()
	}

	if swipe.MessageType.Valid && swipe.MessageType.String == constants.MessageTypeVoice {
		msg.VoiceNoteURL = swipe.VoicenoteURL.Ptr()

		secs, err := utils.NullDecimalToFloatPtr(swipe.MediaSeconds)
		if err != nil {
			return nil, fmt.Errorf("parse swipe media seconds: %w", err)
		}

		msg.MediaSeconds = secs
	}

	return msg, nil
}

// BuildTargetUserFirstMessage reconstructs the liker's first conversation message from their swipe.
// seed is false when legacy voice swipes have no stored duration and no text to fall back to.
func BuildTargetUserFirstMessage(swipe *entity.Swipe, convoID, senderID string) (domain.Message, bool, error) {
	if swipe == nil || !swipe.Message.Valid || !swipe.MessageType.Valid || !swipe.IdempotencyKey.Valid {
		return domain.Message{}, false, nil
	}

	messageType := domain.MessageTypeText

	var textBody *string

	var mediaURL *string

	var mediaSeconds *float64

	if swipe.MessageType.String == constants.MessageTypeVoice {
		secs, err := utils.NullDecimalToFloatPtr(swipe.MediaSeconds)
		if err != nil {
			return domain.Message{}, false, fmt.Errorf("parse swipe media seconds: %w", err)
		}

		if secs == nil {
			if swipe.Message.Valid {
				textBody = swipe.Message.Ptr()
			} else {
				return domain.Message{}, false, nil
			}
		} else {
			messageType = domain.MessageTypeVoice
			mediaURL = swipe.VoicenoteURL.Ptr()
			mediaSeconds = secs
		}
	} else {
		textBody = swipe.Message.Ptr()
	}

	return domain.Message{
		ConversationID: convoID,
		SenderID:       senderID,
		Type:           messageType,
		TextBody:       textBody,
		MediaUrl:       mediaURL,
		MediaSeconds:   mediaSeconds,
		ClientMsgID:    swipe.IdempotencyKey.String,
	}, true, nil
}
