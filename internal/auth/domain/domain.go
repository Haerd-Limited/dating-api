package domain

import (
	"time"
)

type (
	RefreshToken struct {
		ID        string // uuid
		UserID    string
		Token     string // secure random string
		ExpiresAt time.Time
	}
	Register struct {
		Email       string
		PhoneNumber string
		FirstName   string
		LastName    *string
	}

	Login struct {
		PhoneNumber string
	}

	Refresh struct {
		RefreshToken string
	}

	RevokeRefreshToken struct {
		RefreshToken string
	}
	AuthTokensAndUserID struct {
		UserID       string
		RefreshToken string
		AccessToken  string
	}
)
