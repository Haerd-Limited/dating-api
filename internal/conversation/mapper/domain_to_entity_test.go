package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

func TestMapMessageDomainToEntityUsesSharedDecimalHelper(t *testing.T) {
	secs := 3.5
	text := "hi"

	msg := MapMessageDomainToEntity(domain.Message{
		ConversationID: "convo-1",
		SenderID:       "user-1",
		Type:           domain.MessageTypeVoice,
		TextBody:       &text,
		MediaUrl:       strPtr("https://example.com/voice.m4a"),
		MediaSeconds:   &secs,
		ClientMsgID:    "client-1",
	})

	require.False(t, msg.MediaSeconds.IsZero())

	got, err := utils.NullDecimalToFloatPtr(msg.MediaSeconds)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.InDelta(t, secs, *got, 0.001)
	assert.Equal(t, string(domain.MessageTypeVoice), msg.Type)
	assert.Equal(t, constants.MessageTypeVoice, msg.Type)
}

func strPtr(s string) *string {
	return &s
}
