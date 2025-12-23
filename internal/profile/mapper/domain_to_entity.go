package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

func MapProfileToEntityForUpdate(p *domain.Profile) (*entity.UserProfile, []string, error) {
	if p == nil {
		return nil, nil, nil
	}

	var columnWhitelist []string

	ent := &entity.UserProfile{}

	if p.UserID != "" {
		ent.UserID = p.UserID
	}

	if p.VerifiedStatus != "" {
		ent.Verified = p.VerifiedStatus

		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Verified)
	}

	if p.Emoji != "" && p.Emoji != constants.DefaultEmoji {
		ent.Emoji = null.StringFrom(p.Emoji)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Emoji)
	}

	if p.CoverMediaURL != nil {
		ent.CoverMediaURL = null.StringFromPtr(p.CoverMediaURL)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.CoverMediaURL)
	}

	if p.CoverMediaType != nil {
		ent.CoverMediaType = null.StringFromPtr(p.CoverMediaType)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.CoverMediaType)
	}

	if p.CoverMediaAspectRatio != nil {
		aspectRatio32 := float32(*p.CoverMediaAspectRatio)
		ent.CoverMediaAspectRatio = null.Float32FromPtr(&aspectRatio32)

		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.CoverMediaAspectRatio)
	}

	// Strings
	if p.DisplayName != "" {
		ent.DisplayName = p.DisplayName
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.DisplayName)
	}

	if p.City != "" {
		ent.City = null.StringFrom(p.City)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.City)
	}

	if p.Country != "" {
		ent.Country = null.StringFrom(p.Country)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Country)
	}

	if p.Work != nil {
		ent.Work = null.StringFrom(*p.Work)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Work)
	}

	if p.JobTitle != nil {
		ent.JobTitle = null.StringFrom(*p.JobTitle)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.JobTitle)
	}

	if p.University != nil {
		ent.University = null.StringFrom(*p.University)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.University)
	}

	if !p.Birthdate.IsZero() {
		ent.Birthdate = null.TimeFrom(p.Birthdate)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Birthdate)
	}

	// Scalars
	if p.HeightCM != 0 {
		ent.HeightCM = null.Int16From(p.HeightCM)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.HeightCM)
	}

	// FKs (SMALLINT → int16)
	if p.GenderID != 0 {
		ent.GenderID = null.Int16From(p.GenderID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.GenderID)
	}

	if p.DatingIntentionID != 0 {
		ent.DatingIntentionID = null.Int16From(p.DatingIntentionID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.DatingIntentionID)
	}

	if p.SexualityID != 0 {
		ent.SexualityID = null.Int16From(p.SexualityID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.SexualityID)
	}

	if p.ReligionID != 0 {
		ent.ReligionID = null.Int16From(p.ReligionID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.ReligionID)
	}

	if p.EducationLevelID != 0 {
		ent.EducationLevelID = null.Int16From(p.EducationLevelID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.EducationLevelID)
	}

	if p.PoliticalBeliefID != 0 {
		ent.PoliticalBeliefID = null.Int16From(p.PoliticalBeliefID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.PoliticalBeliefID)
	}

	if p.DrinkingID != 0 {
		ent.DrinkingID = null.Int16From(p.DrinkingID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.DrinkingID)
	}

	if p.SmokingID != 0 {
		ent.SmokingID = null.Int16From(p.SmokingID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.SmokingID)
	}

	if p.MarijuanaID != 0 {
		ent.MarijuanaID = null.Int16From(p.MarijuanaID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.MarijuanaID)
	}

	if p.DrugsID != 0 {
		ent.DrugsID = null.Int16From(p.DrugsID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.DrugsID)
	}

	if p.ChildrenStatusID != nil {
		ent.ChildrenStatusID = null.Int16FromPtr(p.ChildrenStatusID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.ChildrenStatusID)
	}

	if p.FamilyPlanID != nil {
		ent.FamilyPlanID = null.Int16From(int16(*p.FamilyPlanID))
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.FamilyPlanID)
	}

	// EthnicityIDs are handled separately via repository methods

	// JSONB: your entity expects []byte
	if p.ProfileMeta != nil {
		b, err := json.Marshal(*p.ProfileMeta)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal profile_meta: %w", err)
		}

		ent.ProfileMeta = null.JSONFrom(b)
	}

	// Location → Geo
	if p.Latitude != 0.0 && p.Longitude != 0.0 {
		ent.Geo = fmt.Sprintf("SRID=4326;POINT(%f %f)", p.Longitude, p.Latitude)

		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Geo)
	}

	return ent, columnWhitelist, nil
}

func MapUpdatedPhotosToEntity(updatedPhotos []domain.Photo, userID string) []entity.Photo {
	var out []entity.Photo
	for _, p := range updatedPhotos {
		out = append(out, entity.Photo{
			UserID:    null.StringFrom(userID),
			URL:       p.URL,
			Position:  null.Int16From(p.Position),
			IsPrimary: p.IsPrimary,
		})
	}

	return out
}

func MapVoicePromptsUpdateToEntity(uploadedPrompts []domain.VoicePromptUpdate, userID string) []entity.VoicePrompt {
	var out []entity.VoicePrompt

	for _, p := range uploadedPrompts {
		vp := entity.VoicePrompt{
			UserID:     null.StringFrom(userID),
			PromptType: null.Int16From(p.PromptTypeID),
			Position:   null.Int16From(p.Position),
			IsPrimary:  p.IsPrimary,
			AudioURL:   p.URL,
			DurationMS: p.DurationMs,
		}
		if p.PromptCoverMediaURL != "" {
			vp.CoverMediaURL = null.StringFrom(p.PromptCoverMediaURL)
		}

		if p.PromptCoverMediaType != nil {
			vp.CoverMediaType = null.StringFromPtr(p.PromptCoverMediaType)
		}

		if p.PromptCoverMediaAspectRatio != nil {
			aspectRatio32 := float32(*p.PromptCoverMediaAspectRatio)
			vp.CoverMediaAspectRatio = null.Float32FromPtr(&aspectRatio32)
		}

		out = append(out, vp)
	}

	return out
}
