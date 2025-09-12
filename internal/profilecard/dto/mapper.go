package dto

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/profilecard"
)

func ProfileCardsToDto(profiles []profilecard.ProfileCard) []ProfileCard {
	var result []ProfileCard
	for _, profile := range profiles {
		result = append(result, ProfileCardToDto(profile))
	}

	return result
}

func ProfileCardToDto(profile profilecard.ProfileCard) ProfileCard {
	// Format times
	birthdateStr := profile.Birthdate.Format(time.DateOnly)
	createdAtStr := profile.CreatedAt.Format(time.RFC3339)
	updatedAtStr := profile.UpdatedAt.Format(time.RFC3339)

	// Map voice prompts
	var voicePrompts []VoicePrompt
	for _, vp := range profile.VoicePrompts {
		voicePrompts = append(voicePrompts, VoicePrompt{
			URL: vp.URL,
			PromptType: Prompt{
				ID:       vp.PromptType.ID,
				Key:      vp.PromptType.Key,
				Label:    vp.PromptType.Label,
				Category: vp.PromptType.Category,
			},
			IsPrimary:  vp.IsPrimary,
			Position:   vp.Position,
			DurationMs: vp.DurationMs,
		})
	}

	return ProfileCard{
		DisplayName: profile.DisplayName,
		Birthdate:   birthdateStr,
		Age:         profile.Age,
		HeightCM:    profile.HeightCM,
		UserID:      profile.UserID,

		Latitude:  profile.Latitude,
		Longitude: profile.Longitude,
		City:      profile.City,
		Country:   profile.Country,

		Gender:          profile.Gender,
		DatingIntention: profile.DatingIntention,
		Religion:        profile.Religion,
		EducationLevel:  profile.EducationLevel,
		PoliticalBelief: profile.PoliticalBelief,
		Drinking:        profile.Drinking,
		Smoking:         profile.Smoking,
		Marijuana:       profile.Marijuana,
		Drugs:           profile.Drugs,
		ChildrenStatus:  profile.ChildrenStatus,
		FamilyPlan:      profile.FamilyPlan,
		Ethnicity:       profile.Ethnicity,
		SpokenLanguages: profile.SpokenLanguages,
		VoicePrompts:    voicePrompts,
		Verified:        profile.Verified,
		Theme: UserTheme{
			BaseHex: profile.Theme.BaseHex,
			Palette: profile.Theme.Palette,
		},

		Work:       profile.Work,
		JobTitle:   profile.JobTitle,
		University: profile.University,

		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}
