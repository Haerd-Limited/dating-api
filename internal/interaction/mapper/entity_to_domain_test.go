package mapper

import (
	"testing"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	interactiondomain "github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

func TestSwipeToEntitySetsMediaSeconds(t *testing.T) {
	secs := 8.5
	msg := "hello"
	msgType := constants.MessageTypeVoice
	voiceURL := "https://example.com/voice.m4a"

	entitySwipe := SwipeToEntity(interactiondomain.Swipe{
		UserID:       "user-1",
		TargetUserID: "user-2",
		Action:       constants.ActionLike,
		Message:      &msg,
		MessageType:  &msgType,
		VoiceNoteURL: &voiceURL,
		MediaSeconds: &secs,
	})

	require.False(t, entitySwipe.MediaSeconds.IsZero())

	got, err := utils.NullDecimalToFloatPtr(entitySwipe.MediaSeconds)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.InDelta(t, secs, *got, 0.001)
}

func TestSwipeMessageToDomainVoiceLike(t *testing.T) {
	msg := "nice prompt"
	msgType := constants.MessageTypeVoice
	voiceURL := "https://example.com/voice.m4a"
	secs := 12.5

	swipe := &entity.Swipe{
		Message:      null.StringFrom(msg),
		MessageType:  null.StringFrom(msgType),
		VoicenoteURL: null.StringFrom(voiceURL),
		MediaSeconds: utils.FloatPtrToNullDecimal(&secs),
	}

	got, err := SwipeMessageToDomain(swipe)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.MessageText)
	assert.Equal(t, msg, *got.MessageText)
	require.NotNil(t, got.MessageType)
	assert.Equal(t, msgType, *got.MessageType)
	require.NotNil(t, got.VoiceNoteURL)
	assert.Equal(t, voiceURL, *got.VoiceNoteURL)
	require.NotNil(t, got.MediaSeconds)
	assert.InDelta(t, secs, *got.MediaSeconds, 0.001)
}

func TestSwipeMessageToDomainTextLike(t *testing.T) {
	msg := "hey"
	msgType := constants.MessageTypeText

	swipe := &entity.Swipe{
		Message:     null.StringFrom(msg),
		MessageType: null.StringFrom(msgType),
	}

	got, err := SwipeMessageToDomain(swipe)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.VoiceNoteURL)
	assert.Nil(t, got.MediaSeconds)
}

func TestBuildTargetUserFirstMessageVoiceWithDuration(t *testing.T) {
	msg := "hello"
	msgType := constants.MessageTypeVoice
	voiceURL := "https://example.com/voice.m4a"
	secs := 4.25

	swipe := &entity.Swipe{
		Message:        null.StringFrom(msg),
		MessageType:    null.StringFrom(msgType),
		VoicenoteURL:   null.StringFrom(voiceURL),
		MediaSeconds:   utils.FloatPtrToNullDecimal(&secs),
		IdempotencyKey: null.StringFrom("client-msg-1"),
	}

	got, seed, err := BuildTargetUserFirstMessage(swipe, "convo-1", "liker-1")
	require.NoError(t, err)
	require.True(t, seed)
	assert.Equal(t, domain.MessageTypeVoice, got.Type)
	require.NotNil(t, got.MediaUrl)
	assert.Equal(t, voiceURL, *got.MediaUrl)
	require.NotNil(t, got.MediaSeconds)
	assert.InDelta(t, secs, *got.MediaSeconds, 0.001)
}

func TestBuildTargetUserFirstMessageLegacyVoiceFallsBackToText(t *testing.T) {
	msg := "legacy caption"
	msgType := constants.MessageTypeVoice

	swipe := &entity.Swipe{
		Message:        null.StringFrom(msg),
		MessageType:    null.StringFrom(msgType),
		VoicenoteURL:   null.StringFrom("https://example.com/voice.m4a"),
		IdempotencyKey: null.StringFrom("client-msg-1"),
	}

	got, seed, err := BuildTargetUserFirstMessage(swipe, "convo-1", "liker-1")
	require.NoError(t, err)
	require.True(t, seed)
	assert.Equal(t, domain.MessageTypeText, got.Type)
	require.NotNil(t, got.TextBody)
	assert.Equal(t, msg, *got.TextBody)
	assert.Nil(t, got.MediaUrl)
	assert.Nil(t, got.MediaSeconds)
}

func TestBuildTargetUserFirstMessageLegacyVoiceWithoutTextSkipsSeed(t *testing.T) {
	msgType := constants.MessageTypeVoice

	swipe := &entity.Swipe{
		MessageType:    null.StringFrom(msgType),
		VoicenoteURL:   null.StringFrom("https://example.com/voice.m4a"),
		IdempotencyKey: null.StringFrom("client-msg-1"),
	}

	_, seed, err := BuildTargetUserFirstMessage(swipe, "convo-1", "liker-1")
	require.NoError(t, err)
	assert.False(t, seed)
}
