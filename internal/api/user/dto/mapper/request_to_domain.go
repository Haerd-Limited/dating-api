package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

func UpdateProfileRequestToDomain(request *validators.UpdateProfileForm, userID string) (*domain.UpdateUserProfile, error) {
	dob, err := time.Parse("2006-01-02", request.DateOfBirth)
	if err != nil {
		return nil, commonErrors.ErrInvalidDob
	}

	if request.Gender != "male" && request.Gender != "female" {
		return nil, commonErrors.ErrInvalidGender
	}

	return &domain.UpdateUserProfile{
		UserID:       userID,
		FullName:     request.FullName,
		Username:     request.Username,
		Email:        request.Email,
		DateOfBirth:  &dob,
		Biography:    request.Bio,
		Gender:       request.Gender,
		ProfileImage: request.ProfileImage,
		ImageHeader:  request.ImageHeader,
	}, nil
}
