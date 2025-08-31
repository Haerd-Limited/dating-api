package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	authDomain "github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

func ToAuthResponse(authResult *authDomain.AuthResult, message string) *dto.AuthResponse {
	if authResult == nil {
		return &dto.AuthResponse{
			Message: message,
		}
	}

	return &dto.AuthResponse{
		Message:      message,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		User: &dto.User{
			ID:             authResult.User.ID,
			OnboardingStep: string(authResult.User.OnboardingStep),
		},
	}
}

func ToRequestCodeResponse(sentTo string) dto.RequestCodeResponse {
	return dto.RequestCodeResponse{
		SentTo: sentTo,
	}
}
