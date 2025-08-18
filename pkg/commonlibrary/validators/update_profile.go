package validators

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/mail"
	"time"
)

type UpdateProfileForm struct {
	FullName    string
	Username    string
	Email       string
	DateOfBirth string // Expected format: YYYY-MM-DD
	Bio         string
	Gender      string

	ProfileImage *multipart.File
	ImageHeader  *multipart.FileHeader
}

func DecodeAndValidateUpdateProfileForm(r *http.Request) (*UpdateProfileForm, error) {
	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// return nil, errors.New("invalid form data")
		return nil, err
	}

	form := &UpdateProfileForm{
		FullName:    r.FormValue("full_name"),
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		DateOfBirth: r.FormValue("date_of_birth"),
		Bio:         r.FormValue("bio"),
		Gender:      r.FormValue("gender"),
	}

	// Validate email if provided
	if form.Email != "" {
		if _, err := mail.ParseAddress(form.Email); err != nil {
			return nil, ErrInvalidEmail
		}
	}

	// Validate date of birth format if provided
	if form.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", form.DateOfBirth); err != nil {
			return nil, ErrInvalidDOBFormat
		}
	}

	// Handle optional profile image
	image, imageHeader, err := r.FormFile("profile_image")
	if err == nil {
		form.ProfileImage = &image
		form.ImageHeader = imageHeader
	} else if err != http.ErrMissingFile {
		return nil, fmt.Errorf("failed reading profile image: %w", err)
	}

	return form, nil
}
