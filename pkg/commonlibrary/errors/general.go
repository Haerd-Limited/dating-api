package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDob           = errors.New("invalid date_of_birth format, use YYYY-MM-DD")
	ErrInvalidGender        = errors.New("invalid gender format")
	ErrUnauthorisedDeletion = fmt.Errorf("unauthorised deletion")
	ErrInvalidUUID          = errors.New("invalid uuid format")
)
