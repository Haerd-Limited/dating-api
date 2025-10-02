package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
)

func UpdateProfileRequestToDomain(req dto.UpdateProfileRequest, userID string) (domain.UpdateProfile, error) {
	var birthdate *time.Time

	if req.Birthdate != nil {
		temp, err := time.Parse(time.DateOnly, *req.Birthdate)
		if err != nil {
			return domain.UpdateProfile{}, commonErrors.ErrInvalidDob
		}

		birthdate = &temp
	}

	var voicePrompts []domain.VoicePromptUpdate

	if req.VoicePrompts != nil {
		for _, vp := range req.VoicePrompts {
			voicePrompts = append(voicePrompts, domain.VoicePromptUpdate{
				URL:            vp.URL,
				PromptTypeID:   vp.PromptType,
				IsPrimary:      vp.IsPrimary,
				Position:       vp.Position,
				DurationMs:     vp.DurationMs,
				PromptCoverURL: vp.CoverPhotoURL,
			})
		}
	}

	var photos []domain.Photo

	if req.Photos != nil {
		for _, p := range req.Photos {
			photos = append(photos, domain.Photo{
				URL:       p.URL,
				IsPrimary: p.IsPrimary,
				Position:  p.Position,
			})
		}
	}

	return domain.UpdateProfile{
		DisplayName:       req.DisplayName,
		Birthdate:         birthdate,
		HeightCM:          req.HeightCM,
		UserID:            userID,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		City:              req.City,
		Country:           req.Country,
		GenderID:          req.GenderID,
		DatingIntentionID: req.DatingIntentionID,
		ReligionID:        req.ReligionID,
		EducationLevelID:  req.EducationLevelID,
		PoliticalBeliefID: req.PoliticalBeliefID,
		DrinkingID:        req.DrinkingID,
		SmokingID:         req.SmokingID,
		MarijuanaID:       req.MarijuanaID,
		DrugsID:           req.DrugsID,
		ChildrenStatusID:  req.ChildrenStatusID,
		FamilyPlanID:      req.FamilyPlanID,
		EthnicityID:       req.EthnicityID,
		SpokenLanguages:   req.SpokenLanguages,
		VoicePrompts:      voicePrompts,
		Photos:            photos,
		Work:              req.Work,
		JobTitle:          req.JobTitle,
		University:        req.University,
		CoverPhotoURL:     req.CoverPhotoUrl,
		ProfileEmoji:      req.ProfileEmoji,
	}, nil
}
