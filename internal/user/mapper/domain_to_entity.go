package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func ToUserEntity(user domain.User) *entity.User {
	userEntity := &entity.User{
		Email:          null.StringFrom(user.Email),
		FirstName:      user.FirstName,
		LastName:       null.StringFromPtr(user.LastName),
		Phone:          null.StringFrom(user.PhoneNumber),
		OnboardingStep: user.OnboardingStep,
	}
	if user.ID != "" {
		userEntity.ID = user.ID
	}

	return userEntity
}
