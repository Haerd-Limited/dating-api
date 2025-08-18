package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func ToUserEntity(user domain.User) *entity.User {
	return &entity.User{
		Email:     null.StringFrom(user.Email),
		FirstName: user.FirstName,
		LastName:  null.StringFromPtr(user.LastName),
		Phone:     null.StringFrom(user.PhoneNumber),
	}
}
