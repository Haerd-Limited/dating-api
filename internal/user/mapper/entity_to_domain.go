package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func UserEntityToUserDomain(user *entity.User) *domain.User {
	if user == nil {
		return nil
	}

	userDomain := &domain.User{
		ID:             user.ID,
		Email:          user.Email.String,
		PhoneNumber:    user.Phone.String,
		FirstName:      user.FirstName,
		OnboardingStep: user.OnboardingStep,
	}

	if user.LastName.Valid {
		userDomain.LastName = &user.LastName.String
	}

	// TODO: Uncomment after running migration and regenerating SQLBoiler entities
	// Run: sqlboiler psql
	// if user.HowDidYouHearAboutUs.Valid {
	// 	userDomain.HowDidYouHearAboutUs = &user.HowDidYouHearAboutUs.String
	// }

	return userDomain
}
