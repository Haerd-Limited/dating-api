package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
)

func MapPromptsToProfileVoicePrompts(uploadedPrompts domain.Prompts) []profiledomain.VoicePromptUpdate {
	var out []profiledomain.VoicePromptUpdate

	for _, p := range uploadedPrompts.UploadedPrompts {
		var coverMediaURL string
		if p.CoverMediaURL != nil {
			coverMediaURL = *p.CoverMediaURL
		}

		out = append(out, profiledomain.VoicePromptUpdate{
			PromptTypeID:                p.PromptType,
			Position:                    p.Position,
			IsPrimary:                   p.IsPrimary,
			URL:                         p.URL,
			PromptCoverMediaURL:         coverMediaURL,
			PromptCoverMediaType:        p.CoverMediaType,
			PromptCoverMediaAspectRatio: p.CoverMediaAspectRatio,
			// todo: add duration somehow. ask frontend to send
		})
	}

	return out
}

func MapUploadedPhotosToProfilePhotos(uploadedPhotos domain.UploadedPhotos) []profiledomain.Photo {
	var out []profiledomain.Photo
	for _, p := range uploadedPhotos.Photos {
		out = append(out, profiledomain.Photo{
			URL:       p.URL,
			Position:  p.Position,
			IsPrimary: p.IsPrimary,
		})
	}

	return out
}
