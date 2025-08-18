package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func UserEntityToUserDomain(user *entity.User) *domain.User {
	if user == nil {
		return nil
	}

	var profilePic *string
	if user.ProfilePicURL.Valid {
		profilePic = &user.ProfilePicURL.String
	}

	var bio string
	if user.Bio.Valid {
		bio = user.Bio.String
	}

	var gender string
	if user.Gender.Valid {
		gender = user.Gender.String
	}

	var dob time.Time
	if user.DateOfBirth.Valid {
		dob = user.DateOfBirth.Time
	}

	return &domain.User{
		ID:              user.ID,
		Email:           user.Email,
		HashedPassword:  user.Password,
		Username:        user.Username,
		FullName:        user.FullName,
		ProfileImageURL: profilePic,
		Bio:             bio,
		Gender:          gender,
		Dob:             dob,
	}
}
