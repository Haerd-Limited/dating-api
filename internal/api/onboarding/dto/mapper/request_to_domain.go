package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func ToDomain(userID string, dto dto.UpdateOnboardingRequest) domain.OnboardingUpdate {
	return domain.OnboardingUpdate{
		UserID: userID,
		UserProfile: &domain.UserProfile{
			// Profile
			DisplayName: dto.DisplayName,
			Birthdate:   dto.Birthdate,
			HeightCM:    dto.HeightCM,

			// Location
			Latitude:  dto.Latitude,
			Longitude: dto.Longitude,
			City:      dto.City,
			Country:   dto.Country,

			// Single-selects
			GenderID:          dto.GenderID,
			DatingIntentionID: dto.DatingIntentionID,
			ReligionID:        dto.ReligionID,
			EducationLevelID:  dto.EducationLevelID,
			PoliticalBeliefID: dto.PoliticalBeliefID,
			DrinkingID:        dto.DrinkingID,
			SmokingID:         dto.SmokingID,
			MarijuanaID:       dto.MarijuanaID,
			DrugsID:           dto.DrugsID,
			ChildrenStatusID:  dto.ChildrenStatusID,
			FamilyPlanID:      dto.FamilyPlanID,
			EthnicityID:       dto.EthnicityID,
			University:        dto.University,
			JobTitle:          dto.JobTitle,
			Work:              dto.Work,
			ProfileMeta:       dto.ProfileMeta,
		},

		Preferences: &domain.Preferences{
			DistanceKM:       dto.DistanceKM,
			AgeMin:           dto.AgeMin,
			AgeMax:           dto.AgeMax,
			SeekGenderIDs:    dto.SeekGenderIDs,
			SeekIntentionIDs: dto.SeekIntentionIDs,
			SeekReligionIDs:  dto.SeekReligionIDs,
			SeekPoliticalIDs: dto.SeekPoliticalIDs,
		},

		// Multi-selects
		LanguageIDs: dto.LanguageIDs,
		InterestIDs: dto.InterestIDs,

		// Progress
		BumpOnboardingStep: dto.BumpOnboardingStep,
	}
}
