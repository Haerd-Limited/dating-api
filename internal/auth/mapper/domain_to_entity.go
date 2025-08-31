package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/aarondl/null/v8"
)

func ToRefreshTokenEntity(refreshToken *domain.RefreshToken) *entity.RefreshToken {
	if refreshToken == nil {
		return nil
	}

	return &entity.RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID,
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt,
	}
}

func ToVerificationCodeEntity(vc *domain.VerificationCode) *entity.VerificationCode {
	if vc == nil {
		return nil
	}

	e := &entity.VerificationCode{
		Channel:     vc.Channel,
		Identifier:  vc.Identifier,
		Purpose:     vc.Purpose,
		CodeHash:    vc.CodeHash,
		ExpiresAt:   vc.ExpiresAt,
		MaxAttempts: vc.MaxAttempts,
	}

	if vc.RequestIP != nil {
		e.RequestIP = null.StringFrom(vc.RequestIP.String())
	} else {
		e.RequestIP = null.String{} // represents SQL NULL
	}

	return e
}
