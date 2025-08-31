package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

func MapRequestCodeRequestToDomain(request dto.RequestCodeRequest, ip string) domain.RequestCode {
	if request.Purpose == "" {
		request.Purpose = "login"
	}

	return domain.RequestCode{
		Channel: request.Channel,
		Email:   request.Email,
		Phone:   request.Phone,
		Purpose: request.Purpose,
		IP:      ip,
	}
}

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
