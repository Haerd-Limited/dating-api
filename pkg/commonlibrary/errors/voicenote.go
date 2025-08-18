package errors

import "errors"

var (
	ErrVoiceNoteTooLong  = errors.New("audio file exceeds 1 minute limit")
	ErrVoiceNoteTooShort = errors.New("audio file is too short")
)
