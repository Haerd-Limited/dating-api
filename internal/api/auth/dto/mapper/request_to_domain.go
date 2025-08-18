package mapper

import (
	"errors"
	"strings"
	"unicode"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

const (
	minNameLen = 3
	maxNameLen = 20
)

var (
	ErrNameContainsSpaces = errors.New("name must not contain spaces")
	ErrInvalidNameLength  = errors.New("name must be between 3 and 20 characters")
)

func MapRegisterRequestToDomain(request dto.RegisterRequest) (*domain.Register, error) {
	/*dob, err := time.Parse("2006-01-02", request.DateOfBirth)
	if err != nil {
		return nil, commonErrors.ErrInvalidDob
	}*/
	firstName := strings.TrimSpace(request.FirstName)

	var lastName string
	if request.LastName != nil {
		lastName = strings.TrimSpace(*request.LastName)
	}

	if hasAnySpace(firstName) || hasAnySpace(lastName) {
		return nil, ErrNameContainsSpaces
	}

	// first name length check
	if l := len(firstName); l < minNameLen || l > maxNameLen {
		return nil, ErrInvalidNameLength
	}

	// last name length check
	if l := len(lastName); l < minNameLen || l > maxNameLen {
		return nil, ErrInvalidNameLength
	}

	if !looksLikeEmail(strings.TrimSpace(request.Email)) {
		return nil, validators.ErrInvalidEmail
	}

	return &domain.Register{
		FirstName:   firstName,
		LastName:    &lastName,
		PhoneNumber: normalizePhone(request.PhoneNumber),
		Email:       strings.TrimSpace(request.Email),
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

func looksLikeEmail(s string) bool { return strings.Contains(s, "@") && strings.Contains(s, ".") }
func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	return s // stub
}
