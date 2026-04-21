package mapper

import (
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

func MapEnrichedProfileToProfileCard(ep profiledomain.EnrichedProfile) profilecard.ProfileCard {
	fp := profilecard.ProfileCard{
		DisplayName:    ep.DisplayName,
		Birthdate:      ep.Birthdate,
		Age:            ep.Age,
		HeightCM:       ep.HeightCM,
		UserID:         ep.UserID,
		Emoji:          ep.Emoji,
		VerifiedStatus: ep.VerifiedStatus,

		Latitude:  ep.Latitude,
		Longitude: ep.Longitude,
		City:      ep.City,
		Country:   ep.Country,

		Gender:          ep.Gender.Label,
		DatingIntention: ep.DatingIntention.Label,
		Sexuality:       ep.Sexuality.Label,
		Religion:        ep.Religion.Label,
		EducationLevel:  ep.EducationLevel.Label,
		PoliticalBelief: ep.PoliticalBelief.Label,
		Drinking:        ep.Drinking.Label,
		Smoking:         ep.Smoking.Label,
		Marijuana:       ep.Marijuana.Label,
		Drugs:           ep.Drugs.Label,

		Work:       ep.Work,
		JobTitle:   ep.JobTitle,
		University: ep.University,

		CreatedAt: ep.CreatedAt,
		UpdatedAt: ep.UpdatedAt,
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

	// Ethnicities -> []string (labels)
	if len(ep.Ethnicities) > 0 {
		fp.Ethnicities = make([]string, len(ep.Ethnicities))
		for i, e := range ep.Ethnicities {
			fp.Ethnicities[i] = e.Label
		}
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
				IsPrimary:             vp.IsPrimary,
				Position:              vp.Position,
				DurationMs:            vp.DurationMs,
				WaveformData:          vp.WaveformData,
				CoverMediaURL:         vp.PromptCoverMediaURL,
				CoverMediaType:        vp.PromptCoverMediaType,
				CoverMediaAspectRatio: vp.PromptCoverMediaAspectRatio,
			})
		}
	}

	// Photos
	if len(ep.Photos) > 0 {
		fp.Photos = make([]profilecard.Photo, 0, len(ep.Photos))
		for _, p := range ep.Photos {
			fp.Photos = append(fp.Photos, profilecard.Photo{
				URL:       p.URL,
				IsPrimary: p.IsPrimary,
				Position:  p.Position,
			})
		}
	}

	return fp
}
