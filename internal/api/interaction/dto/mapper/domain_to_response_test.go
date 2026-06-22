package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

func TestMapToGetLikesResponseIncludesVoiceMessageFields(t *testing.T) {
	msgType := constants.MessageTypeVoice
	msgText := "nice voice"
	voiceURL := "https://example.com/voice.m4a"
	secs := 6.0

	resp := MapToGetLikesResponse(&domain.Likes{
		FreeToMatch: []domain.Like{
			{
				Message: &domain.Message{
					MessageText:  &msgText,
					MessageType:  &msgType,
					VoiceNoteURL: &voiceURL,
					MediaSeconds: &secs,
				},
			},
		},
	})

	require.Len(t, resp.FreeToMatch, 1)
	require.NotNil(t, resp.FreeToMatch[0].Message)
	require.NotNil(t, resp.FreeToMatch[0].Message.VoiceNoteURL)
	assert.Equal(t, voiceURL, *resp.FreeToMatch[0].Message.VoiceNoteURL)
	require.NotNil(t, resp.FreeToMatch[0].Message.MediaSeconds)
	assert.InDelta(t, secs, *resp.FreeToMatch[0].Message.MediaSeconds, 0.001)
}
