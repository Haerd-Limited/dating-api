package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func MapIntroRequestToDomain(request dto.IntroRequest, userID string) domain.Intro {
	return domain.Intro{
		UserID:               userID,
		FirstName:            request.FirstName,
		LastName:             request.LastName,
		Email:                request.Email,
		HowDidYouHearAboutUs: request.HowDidYouHearAboutUs,
	}
}

func MapBasicRequestToDomain(req dto.BasicsRequest, userID string) domain.Basics {
	return domain.Basics{
		UserID:            userID,
		Birthdate:         req.Birthdate,
		HeightCm:          req.HeightCm,
		GenderID:          req.GenderID,
		DatingIntentionID: req.DatingIntentionID,
		SexualityID:       req.SexualityID,
	}
}

func MapLocationRequestToDomain(req dto.LocationRequest, userID string) domain.Location {
	return domain.Location{
		UserID:    userID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		City:      req.City,
		Country:   req.Country,
	}
}

func MapLifestyleRequestToDomain(req dto.LifestyleRequest, userID string) domain.Lifestyle {
	return domain.Lifestyle{
		UserID:      userID,
		DrinkingID:  req.DrinkingID,
		MarijuanaID: req.MarijuanaID,
		SmokingID:   req.SmokingID,
		DrugsID:     req.DrugsID,
	}
}

func MapBackgroundRequestToDomain(req dto.BackgroundRequest, userID string) domain.Background {
	return domain.Background{
		UserID:           userID,
		EducationLevelID: req.EducationLevelID,
		EthnicityIDs:     req.EthnicityIDs,
	}
}

func MapBeliefsRequestToDomain(req dto.BeliefsRequest, userID string) domain.Beliefs {
	return domain.Beliefs{
		UserID:             userID,
		PoliticalBeliefsID: req.PoliticalBeliefID,
		ReligionID:         req.ReligionID,
	}
}

func MapLanguagesRequestToDomain(req dto.LanguagesRequest, userID string) domain.Languages {
	return domain.Languages{
		UserID:      userID,
		LanguageIDs: req.LanguageIDs,
	}
}

func MapWorkAndEducationRequestToDomain(req dto.WorkAndEducationRequest, userID string) domain.WorkAndEducation {
	return domain.WorkAndEducation{
		UserID:     userID,
		Workplace:  req.Workplace,
		JobTitle:   req.JobTitle,
		University: req.University,
	}
}

func MapPhotosRequestToDomain(req dto.PhotosRequest, userID string) domain.UploadedPhotos {
	var photos []domain.Photo

	for _, p := range req.UploadedPhotos {
		photos = append(photos, domain.Photo{
			URL:       p.URL,
			Position:  p.Position,
			IsPrimary: p.IsPrimary,
		})
	}

	return domain.UploadedPhotos{
		UserID: userID,
		Photos: photos,
	}
}

func MapProfileToDomain(req dto.ProfileRequest, userID string) domain.Profile {
	return domain.Profile{
		UserID:                       userID,
		ProfileBaseColour:            req.ProfileBaseColour,
		ProfileCoverMediaURL:         req.ProfileCoverMediaURL,
		ProfileCoverMediaType:        req.ProfileCoverMediaType,
		ProfileCoverMediaAspectRatio: req.ProfileCoverMediaAspectRatio,
	}
}

func MapPromptsRequestToDomain(req dto.PromptsRequest, userID string) domain.Prompts {
	var voicePrompts []domain.VoicePrompt

	for _, p := range req.UploadedPrompts {
		voicePrompts = append(voicePrompts, domain.VoicePrompt{
			URL:                   p.URL,
			Position:              p.Position,
			IsPrimary:             p.IsPrimary,
			PromptType:            p.PromptType,
			WaveformData:          p.WaveformData,
			CoverMediaURL:         p.CoverMediaURL,
			CoverMediaType:        p.CoverMediaType,
			CoverMediaAspectRatio: p.CoverMediaAspectRatio,
		})
	}

	return domain.Prompts{
		UserID:          userID,
		UploadedPrompts: voicePrompts,
	}
}

func MapVideoVerificationRequestToDomain(req dto.VideoVerificationRequest, userID string) domain.VideoVerification {
	return domain.VideoVerification{
		UserID:   userID,
		VideoKey: req.VideoKey,
	}
}
