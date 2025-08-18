package audio

import (
	"errors"
	"mime/multipart"
	"time"

	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
)

const (
	MaxVoiceNoteDuration = time.Minute
	MinVoiceNoteDuration = 5 * time.Second
)

func ValidateVoiceNoteDuration(file *multipart.File) (int, error) {
	if file == nil {
		return 0, errors.New("file is nil")
	}

	duration, err := GetAudioDuration(file)
	if err != nil {
		return 0, err
	}

	if duration > MaxVoiceNoteDuration {
		return 0, commonErrors.ErrVoiceNoteTooLong
	}

	if duration < MinVoiceNoteDuration {
		return 0, commonErrors.ErrVoiceNoteTooShort
	}

	result := int(duration.Seconds())

	return result, nil
}
