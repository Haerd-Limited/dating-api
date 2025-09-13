package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
)

func MapProfileToEntity(p *domain.Profile) (*entity.UserProfile, error) {
	if p == nil {
		return nil, nil
	}

	ent := &entity.UserProfile{}

	if p.UserID != "" {
		ent.UserID = p.UserID
	}

	// Strings
	ent.DisplayName = p.DisplayName

	if p.City != "" {
		ent.City = null.StringFrom(p.City)
	}

	if p.Country != "" {
		ent.Country = null.StringFrom(p.Country)
	}

	if p.Work != nil {
		ent.Work = null.StringFrom(*p.Work)
	}

	if p.JobTitle != nil {
		ent.JobTitle = null.StringFrom(*p.JobTitle)
	}

	if p.University != nil {
		ent.University = null.StringFrom(*p.University)
	}

	ent.Birthdate = null.TimeFrom(p.Birthdate)

	// Scalars
	ent.HeightCM = null.Int16From(p.HeightCM)

	// FKs (SMALLINT → int16)
	ent.GenderID = null.Int16From(p.GenderID)

	ent.DatingIntentionID = null.Int16From(p.DatingIntentionID)

	if p.ReligionID != 0 {
		ent.ReligionID = null.Int16From(p.ReligionID)
	}

	if p.EducationLevelID != 0 {
		ent.EducationLevelID = null.Int16From(p.EducationLevelID)
	}

	if p.PoliticalBeliefID != 0 {
		ent.PoliticalBeliefID = null.Int16From(p.PoliticalBeliefID)
	}

	if p.DrinkingID != 0 {
		ent.DrinkingID = null.Int16From(p.DrinkingID)
	}

	if p.SmokingID != 0 {
		ent.SmokingID = null.Int16From(p.SmokingID)
	}

	if p.MarijuanaID != 0 {
		ent.MarijuanaID = null.Int16From(p.MarijuanaID)
	}

	if p.DrugsID != 0 {
		ent.DrugsID = null.Int16From(p.DrugsID)
	}

	if p.ChildrenStatusID != nil {
		ent.ChildrenStatusID = null.Int16FromPtr(p.ChildrenStatusID)
	}

	if p.FamilyPlanID != nil {
		ent.FamilyPlanID = null.Int16From(*p.FamilyPlanID)
	}

	if p.EthnicityID != 0 {
		ent.EthnicityID = null.Int16From(p.EthnicityID)
	}

	// JSONB: your entity expects []byte
	if p.ProfileMeta != nil {
		b, err := json.Marshal(*p.ProfileMeta)
		if err != nil {
			return nil, fmt.Errorf("marshal profile_meta: %w", err)
		}

		ent.ProfileMeta = null.JSONFrom(b)
	}

	// NOTE: lat/lon → geo is handled in the repository with ST_MakePoint if you decide to pass them down.

	// Location → Geo
	if p.Latitude != 0.0 && p.Longitude != 0.0 {
		ent.Geo = fmt.Sprintf("SRID=4326;POINT(%f %f)", p.Longitude, p.Latitude)
	} else {
		// fallback default to avoid PostGIS parse error
		ent.Geo = "SRID=4326;POINT(0 0)"
	}

	return ent, nil
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

func MapVoicePromptsToEntity(uploadedPrompts []domain.VoicePrompt, userID string) []entity.VoicePrompt {
	var out []entity.VoicePrompt
	for _, p := range uploadedPrompts {
		out = append(out, entity.VoicePrompt{
			UserID:     null.StringFrom(userID),
			PromptType: null.Int16From(p.PromptType.ID),
			Position:   null.Int16From(p.Position),
			IsPrimary:  p.IsPrimary,
			AudioURL:   p.URL,
			// todo: add transcript
			// todo: add duration somehow
		})
	}

	return out
}
