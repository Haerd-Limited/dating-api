package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/media/dto"
	"github.com/Haerd-Limited/dating-api/internal/media/domain"
)

func MapUploadURLToResponse(url domain.UploadUrl) dto.UploadUrl {
	return dto.UploadUrl{
		Key:       url.Key,
		UploadUrl: url.UploadUrl,
		Headers:   url.Headers,
		MaxBytes:  url.MaxBytes,
	}
}
