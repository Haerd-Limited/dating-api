package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func MapPromptsToEntity(uploadedPrompts domain.Prompts) []entity.VoicePrompt {
	var out []entity.VoicePrompt
	for _, p := range uploadedPrompts.UploadedPrompts {
		out = append(out, entity.VoicePrompt{
			UserID:        null.StringFrom(uploadedPrompts.UserID),
			PromptType:    null.Int16From(p.PromptType),
			Position:      null.Int16From(p.Position),
			IsPrimary:     p.IsPrimary,
			AudioURL:      p.URL,
			CoverPhotoURL: null.StringFromPtr(p.CoverPhotoUrl),
			// todo: add transcript
			// todo: add duration somehow
		})
	}

	return out
}

func MapUploadedPhotosToEntity(uploadedPhotos domain.UploadedPhotos) []entity.Photo {
	var out []entity.Photo
	for _, p := range uploadedPhotos.Photos {
		out = append(out, entity.Photo{
			UserID:    null.StringFrom(uploadedPhotos.UserID),
			URL:       p.URL,
			Position:  null.Int16From(p.Position),
			IsPrimary: p.IsPrimary,
		})
	}

	return out
}
