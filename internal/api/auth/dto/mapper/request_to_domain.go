package mapper

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

const (
	minUsernameLen = 3
	maxUsernameLen = 20
)

var (
	ErrUsernameContainsSpaces = errors.New("username must not contain spaces")
	ErrInvalidUserameLength   = errors.New("username must be between 3 and 20 characters")
)

func MapRegisterRequestToDomain(request *validators.RegisterForm) (*domain.Register, error) {
	dob, err := time.Parse("2006-01-02", request.DateOfBirth)
	if err != nil {
		return nil, commonErrors.ErrInvalidDob
	}

	if request.Gender != "male" && request.Gender != "female" {
		return nil, commonErrors.ErrInvalidGender
	}

	username := strings.TrimSpace(request.Username)

	if hasAnySpace(username) {
		return nil, ErrUsernameContainsSpaces
	}

	// Username length check
	if l := len(username); l < minUsernameLen || l > maxUsernameLen {
		return nil, ErrInvalidUserameLength
	}

	return &domain.Register{
		FullName:           strings.TrimSpace(request.FullName),
		Username:           username,
		Email:              strings.TrimSpace(request.Email),
		Password:           strings.TrimSpace(request.Password),
		DateOfBirth:        dob,
		Bio:                request.Bio,
		Gender:             request.Gender,
		ProfileImage:       &request.ProfileImage,
		ProfileImageHeader: request.ImageHeader,
	}, nil
}

func MapLoginRequestToDomain(request dto.LoginRequest) domain.Login {
	return domain.Login{
		Email:    request.Email,
		Password: request.Password,
	}
}

func MapRefreshRequestToDomain(request dto.RefreshRequest) domain.Refresh {
	return domain.Refresh{
		RefreshToken: request.RefreshToken,
	}
}

func MapLogoutRequestToDomain(request dto.LogoutRequest) domain.RevokeRefreshToken {
	return domain.RevokeRefreshToken{
		RefreshToken: request.RefreshToken,
	}
}

// hasAnySpace returns true if s contains any Unicode whitespace character.
func hasAnySpace(s string) bool {
	return strings.IndexFunc(s, unicode.IsSpace) >= 0
}
