package mapper

import (
	discoverdomain "github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/user/dto"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
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

		Work:       profile.Work,
		JobTitle:   profile.JobTitle,
		University: profile.University,

		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}

func ProfileToDto(profile domain.EnrichedProfile) dto.Profile {
	result := dto.Profile{
		DisplayName: profile.DisplayName,
		Birthdate:   profile.Birthdate.Format(time.DateOnly),
		Age:         profile.Age,
		HeightCM:    profile.HeightCM,
		UserID:      profile.UserID,
		Latitude:    profile.Latitude,
		Longitude:   profile.Longitude,
		City:        profile.City,
		Country:     profile.Country,
		Gender: dto.Gender{
			ID:    profile.Gender.ID,
			Label: profile.Gender.Label,
		},
		DatingIntention: dto.DatingIntention{
			ID:    profile.DatingIntention.ID,
			Label: profile.DatingIntention.Label,
		},
		Religion: dto.Religion{
			ID:    profile.Religion.ID,
			Label: profile.Religion.Label,
		},
		EducationLevel: dto.EducationLevel{
			ID:    profile.EducationLevel.ID,
			Label: profile.EducationLevel.Label,
		},
		PoliticalBelief: dto.PoliticalBelief{
			ID:    profile.PoliticalBelief.ID,
			Label: profile.PoliticalBelief.Label,
		},
		Drinking: dto.Habit{
			ID:    profile.Drinking.ID,
			Label: profile.Drinking.Label,
		},
		Smoking: dto.Habit{
			ID:    profile.Smoking.ID,
			Label: profile.Smoking.Label,
		},
		Marijuana: dto.Habit{
			ID:    profile.Marijuana.ID,
			Label: profile.Marijuana.Label,
		},
		Drugs: dto.Habit{
			ID:    profile.Drugs.ID,
			Label: profile.Drugs.Label,
		},
		Ethnicity: dto.Ethnicity{
			ID:    profile.Ethnicity.ID,
			Label: profile.Ethnicity.Label,
		},
		Work:       profile.Work,
		JobTitle:   profile.JobTitle,
		University: profile.University,
		CreatedAt:  profile.CreatedAt.Format(time.DateOnly),
		UpdatedAt:  profile.UpdatedAt.Format(time.DateOnly),
	}

	if profile.SpokenLanguages != nil {
		for _, language := range profile.SpokenLanguages {
			result.SpokenLanguages = append(result.SpokenLanguages, dto.Language{
				ID:    language.ID,
				Label: language.Label,
			})
		}
	}

	if profile.VoicePrompts != nil {
		for _, prompt := range profile.VoicePrompts {
			result.VoicePrompts = append(result.VoicePrompts, dto.VoicePrompt{
				URL: prompt.URL,
				PromptType: dto.Prompt{
					ID:       prompt.PromptType.ID,
					Label:    prompt.PromptType.Label,
					Key:      prompt.PromptType.Key,
					Category: prompt.PromptType.Category,
				},
				IsPrimary:  prompt.IsPrimary,
				Position:   prompt.Position,
				DurationMs: prompt.DurationMs,
			})
		}
	}

	if profile.Photos != nil {
		for _, photo := range profile.Photos {
			result.Photos = append(result.Photos, dto.Photo{
				URL:       photo.URL,
				IsPrimary: photo.IsPrimary,
				Position:  photo.Position,
			})
		}
	}

	if profile.ChildrenStatus != nil {
		result.ChildrenStatus = &dto.Status{
			ID:    result.ChildrenStatus.ID,
			Label: result.ChildrenStatus.Label,
			Key:   result.ChildrenStatus.Key,
		}
	}

	if profile.FamilyPlan != nil {
		result.FamilyPlan = &dto.Status{
			ID:    profile.FamilyPlan.ID,
			Label: profile.FamilyPlan.Label,
			Key:   profile.FamilyPlan.Key,
		}
	}

	return result
}
