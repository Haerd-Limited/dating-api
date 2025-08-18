package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func ToUserEntity(user *domain.User) *entity.User {
	if user == nil {
		return nil
	}

	return &entity.User{
		FullName:      user.FullName,
		Username:      user.Username,
		Email:         user.Email,
		Password:      user.HashedPassword,
		DateOfBirth:   null.TimeFrom(user.Dob),
		Gender:        null.StringFrom(user.Gender),
		Bio:           null.StringFrom(user.Bio),
		ProfilePicURL: null.StringFromPtr(user.ProfileImageURL),
	}
}
