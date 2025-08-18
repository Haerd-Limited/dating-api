package domain

import (
	"mime/multipart"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

type (
	RefreshToken struct {
		ID        string // uuid
		UserID    string
		Token     string // secure random string
		ExpiresAt time.Time
	}
	Register struct {
		FullName           string
		Username           string
		Email              string
		Password           string
		DateOfBirth        time.Time
		Bio                string
		Gender             string
		ProfileImage       *multipart.File
		ProfileImageHeader *multipart.FileHeader
	}

	Login struct {
		Email    string
		Password string
	}

	Refresh struct {
		RefreshToken string
	}

	RevokeRefreshToken struct {
		RefreshToken string
	}

	AuthTokensAndUser struct {
		RefreshToken string
		AccessToken  string
		User         domain.User
	}
	AuthTokens struct {
		RefreshToken string
		AccessToken  string
	}
)
