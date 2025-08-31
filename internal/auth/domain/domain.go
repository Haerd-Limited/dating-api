package domain

import (
	"net"
	"time"
)

type (
	RefreshToken struct {
		ID        string // uuid
		UserID    string
		Token     string // secure random string
		ExpiresAt time.Time
	}

	RequestCode struct {
		Channel string
		Email   *string
		Phone   *string
		Purpose string
		IP      string
	}

	VerificationCode struct {
		Channel     string
		Identifier  string
		Purpose     string
		CodeHash    string
		ExpiresAt   time.Time
		RequestIP   net.IP
		MaxAttempts int16
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
