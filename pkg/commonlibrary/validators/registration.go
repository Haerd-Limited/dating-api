package validators

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type RegisterForm struct {
	FullName    string
	Username    string
	Email       string
	Password    string
	DateOfBirth string // Expected format: YYYY-MM-DD
	Bio         string
	Gender      string

	ProfileImage multipart.File
	ImageHeader  *multipart.FileHeader
}

var (
	// ErrInvalidDOBFormat is returned when DOB is not YYYY-MM-DD
	ErrInvalidDOBFormat = errors.New("date_of_birth must be in YYYY-MM-DD format")
	// ErrInvalidEmail is returned when email is not a valid email
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrMissingRequiredField = errors.New("all fields except bio and image are required")
	emailRegex              = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func DecodeAndValidateRegisterForm(r *http.Request) (*RegisterForm, error) {
	// Parse the form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	form := &RegisterForm{
		FullName:    r.FormValue("full_name"),
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		Password:    r.FormValue("password"),
		DateOfBirth: r.FormValue("date_of_birth"),
		Bio:         r.FormValue("bio"),
		Gender:      r.FormValue("gender"),
	}

	// Required field checks
	if form.FullName == "" || form.Username == "" || form.Email == "" || form.Password == "" || form.DateOfBirth == "" || form.Gender == "" {
		return nil, ErrMissingRequiredField
	}

	// Validate email
	if !IsValidEmail(strings.TrimSpace(form.Email)) {
		return nil, ErrInvalidEmail
	}

	// Validate Date of Birth format
	if _, err := time.Parse("2006-01-02", form.DateOfBirth); err != nil {
		return nil, ErrInvalidDOBFormat
	}

	// Handle profile image (optional)
	image, imageHeader, err := r.FormFile("profile_image")
	if err == nil {
		form.ProfileImage = image
		form.ImageHeader = imageHeader
	} else if !errors.Is(err, http.ErrMissingFile) {
		return nil, fmt.Errorf("failed reading profile image: %w", err)
	}

	return form, nil
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
