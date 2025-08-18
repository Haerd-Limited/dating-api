package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	authDomain "github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

func ToRegisterResponse(authResult *authDomain.AuthTokensAndUser, message string) *dto.RegisterResponse {
	if authResult == nil {
		return &dto.RegisterResponse{
			Message: message,
		}
	}

	var profilepicURL *string
	if authResult.User.ProfileImageURL != nil {
		profilepicURL = authResult.User.ProfileImageURL
	}

	return &dto.RegisterResponse{
		Message:      message,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		UserDetails: &dto.UserDetails{
			Username:      authResult.User.Username,
			Email:         authResult.User.Email,
			FullName:      authResult.User.FullName,
			ProfilePicURL: profilepicURL,
		},
	}
}

func MapAuthTokensAndUserResponse(authResult *authDomain.AuthTokensAndUser, message string) *dto.LoginResponse {
	if authResult == nil {
		return &dto.LoginResponse{
			Message: message,
		}
	}

	return &dto.LoginResponse{
		Message:      message,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		UserDetails: &dto.UserDetails{
			Username: authResult.User.Username,
			Email:    authResult.User.Email,
			FullName: authResult.User.FullName,
		},
	}
}

func MapAuthTokensToResponse(authResult *authDomain.AuthTokens, message string) *dto.RefreshResponse {
	if authResult == nil {
		return &dto.RefreshResponse{
			Message: message,
		}
	}

	return &dto.RefreshResponse{
		Message:      message,
		RefreshToken: authResult.RefreshToken,
		AccessToken:  authResult.AccessToken,
	}
}
