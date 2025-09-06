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
	// ErrInvalidDOBFormat is returned when DOB is not YYYY-MM-DD
	ErrInvalidDOBFormat = errors.New("date_of_birth must be in YYYY-MM-DD format")
	// ErrInvalidEmail is returned when email is not a valid email
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrMissingRequiredField = errors.New("all fields except bio and image are required")
)
