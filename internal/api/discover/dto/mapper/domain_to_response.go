package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto"
	discoverdomain "github.com/Haerd-Limited/dating-api/internal/discover/domain"
)

func FeedProfilesToDto(profiles []discoverdomain.FeedProfile) []dto.FeedProfile {
	var result []dto.FeedProfile
	for _, profile := range profiles {
		result = append(result, FeedProfileToDto(profile))
	}

	return result
}

func FeedProfileToDto(profile discoverdomain.FeedProfile) dto.FeedProfile {
	// Format times
	birthdateStr := profile.Birthdate.Format(time.DateOnly)
	createdAtStr := profile.CreatedAt.Format(time.RFC3339)
	updatedAtStr := profile.UpdatedAt.Format(time.RFC3339)

	// Map voice prompts
	var voicePrompts []dto.VoicePrompt
	for _, vp := range profile.VoicePrompts {
		voicePrompts = append(voicePrompts, dto.VoicePrompt{
			URL: vp.URL,
			PromptType: dto.Prompt{
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

	return dto.FeedProfile{
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

		Work:       profile.Work,
		JobTitle:   profile.JobTitle,
		University: profile.University,

		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}
