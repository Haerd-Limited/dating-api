package mapper

import (
	"testing"

	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
)

func TestProfileToDtoMapsVoicePromptID(t *testing.T) {
	dtoProfile := ProfileToDto(domain.EnrichedProfile{
		VoicePrompts: []domain.ProfileVoicePrompt{
			{
				ID: 42,
				PromptType: domain.Prompt{
					ID: 7,
				},
			},
		},
	})

	if len(dtoProfile.VoicePrompts) != 1 {
		t.Fatalf("expected 1 voice prompt, got %d", len(dtoProfile.VoicePrompts))
	}

	if dtoProfile.VoicePrompts[0].ID != 42 {
		t.Fatalf("expected voice prompt id 42, got %d", dtoProfile.VoicePrompts[0].ID)
	}

	if dtoProfile.VoicePrompts[0].PromptType.ID != 7 {
		t.Fatalf("expected prompt type id 7, got %d", dtoProfile.VoicePrompts[0].PromptType.ID)
	}
}
