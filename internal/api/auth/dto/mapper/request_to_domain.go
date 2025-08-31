package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

func ToVerifyDomain(req dto.VerifyCodeRequest, ip string) domain.VerifyCode {
	if req.Purpose == "" {
		req.Purpose = "login"
	}

	return domain.VerifyCode{
		Channel: req.Channel,
		Email:   req.Email,
		Phone:   req.Phone,
		Purpose: req.Purpose,
		Code:    req.Code,
		IP:      ip,
	}
}

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
