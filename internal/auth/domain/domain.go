package domain

import (
	"net"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
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

	VerifyCode struct {
		Channel string // "sms" | "email"
		Email   *string
		Phone   *string
		Purpose string // "login" | "register"
		Code    string // 6 digits
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
	AuthResult struct {
		RefreshToken string
		AccessToken  string
		User         *User
	}
	User struct {
		ID             string
		OnboardingStep domain.Steps
	}
	RequestCodeResult struct {
		SentTo         string
		OnboardingStep *domain.Steps // Only set when user exists during registration
	}
)
