package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func MapPromptsToEntity(uploadedPrompts domain.Prompts) []entity.VoicePrompt {
	var out []entity.VoicePrompt
	for _, p := range uploadedPrompts.UploadedPrompts {
		out = append(out, entity.VoicePrompt{
			UserID:        null.StringFrom(uploadedPrompts.UserID),
			PromptType:    null.Int16From(p.PromptType),
			Position:      null.Int16From(p.Position),
			IsPrimary:     p.IsPrimary,
			AudioURL:      p.URL,
			CoverPhotoURL: null.StringFromPtr(p.CoverPhotoUrl),
			// todo: add transcript
			// todo: add duration somehow
		})
	}

	return out
}

func MapUploadedPhotosToEntity(uploadedPhotos domain.UploadedPhotos) []entity.Photo {
	var out []entity.Photo
	for _, p := range uploadedPhotos.Photos {
		out = append(out, entity.Photo{
			UserID:    null.StringFrom(uploadedPhotos.UserID),
			URL:       p.URL,
			Position:  null.Int16From(p.Position),
			IsPrimary: p.IsPrimary,
		})
	}

	return out
}

func MapProfileToEntityForUpdate(p *domain.UserProfile) (*entity.UserProfile, []string, error) {
	if p == nil {
		return nil, nil, nil
	}

	var columnWhitelist []string

	ent := &entity.UserProfile{}

	if p.UserID != "" {
		ent.UserID = p.UserID
	}

	// Strings
	if p.DisplayName != nil {
		ent.DisplayName = *p.DisplayName
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

	if p.ChildrenStatusID != 0 {
		ent.ChildrenStatusID = null.Int16From(p.ChildrenStatusID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.ChildrenStatusID)
	}

	if p.FamilyPlanID != nil {
		ent.FamilyPlanID = null.Int16From(int16(*p.FamilyPlanID))
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.FamilyPlanID)
	}

	if p.EthnicityID != 0 {
		ent.EthnicityID = null.Int16From(p.EthnicityID)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.EthnicityID)
	}

	if p.CoverPhotoURL != nil {
		ent.CoverPhotoURL = null.StringFromPtr(p.CoverPhotoURL)
		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.CoverPhotoURL)
	}

	// JSONB: your entity expects []byte
	if p.ProfileMeta != nil {
		b, err := json.Marshal(*p.ProfileMeta)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal profile_meta: %w", err)
		}

		ent.ProfileMeta = null.JSONFrom(b)
	}

	// NOTE: lat/lon → geo is handled in the repository with ST_MakePoint if you decide to pass them down.

	// Location → Geo
	if p.Latitude != 0.0 && p.Longitude != 0.0 {
		ent.Geo = fmt.Sprintf("SRID=4326;POINT(%f %f)", p.Longitude, p.Latitude)

		columnWhitelist = append(columnWhitelist, entity.UserProfileColumns.Geo)
	}

	return ent, columnWhitelist, nil
}
