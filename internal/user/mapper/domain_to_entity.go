package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func ToUserEntity(user domain.User) *entity.User {
	userEntity := &entity.User{
		ID: user.ID,
	}

	if user.Email != "" {
		userEntity.Email = null.StringFrom(user.Email)
	}

	if user.FirstName != "" {
		userEntity.FirstName = user.FirstName
	}

	if user.LastName != nil {
		userEntity.LastName = null.StringFromPtr(user.LastName)
	}

	if user.PhoneNumber != "" {
		userEntity.Phone = null.StringFrom(user.PhoneNumber)
	}

	if user.OnboardingStep != "" {
		userEntity.OnboardingStep = user.OnboardingStep
	}

	return userEntity
}

func ToUpdatedUserEntity(u domain.User) (*entity.User, []string) {
	e := &entity.User{ID: u.ID}
	cols := []string{}

	if u.Email != "" {
		e.Email = null.StringFrom(u.Email)
		cols = append(cols, entity.UserColumns.Email)
	}

	if u.FirstName != "" {
		e.FirstName = u.FirstName
		cols = append(cols, entity.UserColumns.FirstName)
	}

	if u.LastName != nil { // non-nil => write; "" means empty string, not NULL
		e.LastName = null.StringFromPtr(u.LastName)
		cols = append(cols, entity.UserColumns.LastName)
	}

	if u.PhoneNumber != "" {
		e.Phone = null.StringFrom(u.PhoneNumber)

		cols = append(cols, entity.UserColumns.Phone)
	}

	if u.OnboardingStep != "" {
		e.OnboardingStep = u.OnboardingStep
		cols = append(cols, entity.UserColumns.OnboardingStep)
	}

	if u.HowDidYouHearAboutUs != nil {
		e.HowDidYouHearAboutUs = null.StringFromPtr(u.HowDidYouHearAboutUs)
		cols = append(cols, entity.UserColumns.HowDidYouHearAboutUs)
	}

	return e, cols
}
