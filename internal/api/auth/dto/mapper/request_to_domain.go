package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

func MapLoginRequestToDomain(request dto.LoginRequest) domain.Login {
	return domain.Login{
		PhoneNumber: request.PhoneNumber,
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
