package dto

import (
	"time"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

func ProfileCardsToDto(profiles []profilecard.ProfileCard) []ProfileCard {
	if profiles == nil {
		return []ProfileCard{}
	}

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
			ID:  vp.ID,
			URL: vp.URL,
			PromptType: Prompt{
				ID:       vp.PromptType.ID,
				Key:      vp.PromptType.Key,
				Label:    vp.PromptType.Label,
				Category: vp.PromptType.Category,
			},
			IsPrimary:             vp.IsPrimary,
			Position:              vp.Position,
			DurationMs:            vp.DurationMs,
			CoverMediaURL:         vp.CoverMediaURL,
			CoverMediaType:        vp.CoverMediaType,
			CoverMediaAspectRatio: vp.CoverMediaAspectRatio,
		})
	}

	// Map photos
	var photos []Photo
	for _, p := range profile.Photos {
		photos = append(photos, Photo{
			URL:       p.URL,
			IsPrimary: p.IsPrimary,
			Position:  p.Position,
		})
	}

	var coverMediaURL string
	if profile.CoverMediaURL != nil {
		coverMediaURL = *profile.CoverMediaURL
	}

	return ProfileCard{
		DisplayName: profile.DisplayName,
		Birthdate:   birthdateStr,
		Age:         profile.Age,
		HeightCM:    profile.HeightCM,
		UserID:      profile.UserID,

		Latitude:              profile.Latitude,
		Longitude:             profile.Longitude,
		City:                  profile.City,
		Country:               profile.Country,
		DistanceKm:            profile.DistanceKm,
		Emoji:                 profile.Emoji,
		CoverMediaURL:         coverMediaURL,
		CoverMediaType:        profile.CoverMediaType,
		CoverMediaAspectRatio: profile.CoverMediaAspectRatio,

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
		Ethnicities:     profile.Ethnicities,
		SpokenLanguages: profile.SpokenLanguages,
		VoicePrompts:    voicePrompts,
		Photos:          photos,
		LikeCount:       profile.LikeCount,
		VerifiedStatus:  profile.VerifiedStatus,
		Theme: UserTheme{
			BaseHex: profile.Theme.BaseHex,
			Palette: profile.Theme.Palette,
		},

		Work:         profile.Work,
		JobTitle:     profile.JobTitle,
		University:   profile.University,
		MatchSummary: MapCompatibilitySummary(profile.CompatibilitySummary),

		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}

func MapCompatibilitySummary(ms *profilecard.CompatibilitySummary) *MatchSummary {
	if ms == nil {
		return nil
	}

	out := &MatchSummary{
		MatchPercent: ms.CompatibilityPercent,
		OverlapCount: ms.OverlapCount,
		HiddenReason: ms.HiddenReason,
	}
	if len(ms.Badges) > 0 {
		out.Badges = make([]MatchBadge, 0, len(ms.Badges))
		for _, b := range ms.Badges {
			out.Badges = append(out.Badges, MatchBadge{
				QuestionID:    b.QuestionID,
				QuestionText:  b.QuestionText,
				PartnerAnswer: b.PartnerAnswer,
				Weight:        b.Weight,
				IsMismatch:    b.IsMismatch,
				RequirementBy:  b.RequirementBy,
			})
		}
	}

	return out
}
