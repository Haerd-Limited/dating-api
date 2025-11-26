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

	result := &dto.AuthResponse{
		Message:      message,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
	}

	if authResult.User != nil {
		result.User = &dto.User{
			ID:             authResult.User.ID,
			OnboardingStep: string(authResult.User.OnboardingStep),
		}
	}

	return result
}

func ToRequestCodeResponse(result *authDomain.RequestCodeResult) dto.RequestCodeResponse {
	resp := dto.RequestCodeResponse{
		SentTo: result.SentTo,
	}

	if result.OnboardingStep != nil {
		stepStr := string(*result.OnboardingStep)
		resp.OnboardingStep = &stepStr
	}

	return resp
}
