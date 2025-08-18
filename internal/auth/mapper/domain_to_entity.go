package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
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
