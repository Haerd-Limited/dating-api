package mapper

import (
	"fmt"
	"strings"

	"github.com/Haerd-Limited/dating-api/internal/aws/domain"
)

func MapStringToS3AudioFolderPath(s string) (domain.S3AudioFolderPath, error) {
	switch {
	case strings.EqualFold(s, string(domain.FolderVoiceNotes)):
		return domain.FolderVoiceNotes, nil
	default:
		return "", fmt.Errorf("unknown S3 audio folder: %s", s)
	}
}

func MapStringToS3ImageFolderPath(s string) (domain.S3ImageFolderPath, error) {
	switch {
	case strings.EqualFold(s, string(domain.FolderProfilePictures)):
		return domain.FolderProfilePictures, nil
	case strings.EqualFold(s, string(domain.FolderPostImages)):
		return domain.FolderPostImages, nil
	default:
		return "", fmt.Errorf("unknown S3 image folder: %s", s)
	}
}
