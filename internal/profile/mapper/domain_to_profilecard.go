package mapper

import (
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

func MapEnrichedProfileToProfileCard(ep profiledomain.EnrichedProfile) profilecard.ProfileCard {
	fp := profilecard.ProfileCard{
		DisplayName:   ep.DisplayName,
		Birthdate:     ep.Birthdate,
		Age:           ep.Age,
		HeightCM:      ep.HeightCM,
		UserID:        ep.UserID,
		CoverPhotoUrl: ep.CoverPhotoURL,
		Emoji:         ep.Emoji,

		Latitude:  ep.Latitude,
		Longitude: ep.Longitude,
		City:      ep.City,
		Country:   ep.Country,

		Gender:          ep.Gender.Label,
		DatingIntention: ep.DatingIntention.Label,
		Religion:        ep.Religion.Label,
		EducationLevel:  ep.EducationLevel.Label,
		PoliticalBelief: ep.PoliticalBelief.Label,
		Drinking:        ep.Drinking.Label,
		Smoking:         ep.Smoking.Label,
		Marijuana:       ep.Marijuana.Label,
		Drugs:           ep.Drugs.Label,
		Ethnicity:       ep.Ethnicity.Label,

		Work:       ep.Work,
		JobTitle:   ep.JobTitle,
		University: ep.University,

		CreatedAt: ep.CreatedAt,
		UpdatedAt: ep.UpdatedAt,
		Theme: profilecard.UserTheme{
			BaseHex: ep.Theme.BaseHex,
			Palette: ep.Theme.Palette,
		},
	}

	// Optional statuses -> pointers to label
	if ep.ChildrenStatus != nil {
		lbl := ep.ChildrenStatus.Label
		fp.ChildrenStatus = &lbl
	}

	if ep.FamilyPlan != nil {
		lbl := ep.FamilyPlan.Label
		fp.FamilyPlan = &lbl
	}

	// Spoken languages -> []string (labels)
	if len(ep.SpokenLanguages) > 0 {
		fp.SpokenLanguages = make([]string, len(ep.SpokenLanguages))
		for i, l := range ep.SpokenLanguages {
			fp.SpokenLanguages[i] = l.Label
		}
	}

	// Voice prompts
	if len(ep.VoicePrompts) > 0 {
		fp.VoicePrompts = make([]profilecard.VoicePrompt, 0, len(ep.VoicePrompts))
		for _, vp := range ep.VoicePrompts {
			fp.VoicePrompts = append(fp.VoicePrompts, profilecard.VoicePrompt{
				ID:  vp.ID,
				URL: vp.URL,
				PromptType: profilecard.Prompt{
					ID:       vp.PromptType.ID,
					Key:      vp.PromptType.Key,
					Label:    vp.PromptType.Label,
					Category: vp.PromptType.Category,
				},
				IsPrimary:     vp.IsPrimary,
				Position:      vp.Position,
				DurationMs:    vp.DurationMs,
				CoverPhotoUrl: vp.PromptCoverURL,
			})
		}
	}

	return fp
}
