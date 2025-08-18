package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func UserEntityToUserDomain(user *entity.User) *domain.User {
	if user == nil {
		return nil
	}

	return &domain.User{
		ID:             user.ID,
		Email:          user.Email.String,
		PhoneNumber:    user.Phone.String,
		FirstName:      user.FirstName,
		LastName:       &user.LastName.String,
		OnboardingStep: int(user.OnboardingStep),
	}
}
