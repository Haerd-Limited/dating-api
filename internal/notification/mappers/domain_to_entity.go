package mappers

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/notification/domain"
)

func ToEntity(dt domain.DeviceToken) *entity.DeviceToken {
	return &entity.DeviceToken{
		UserID: dt.UserID,
		Token:  dt.Token,
	}
}
