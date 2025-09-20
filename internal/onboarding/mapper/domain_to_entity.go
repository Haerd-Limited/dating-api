package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

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

func MapProfileToEntity(p *domain.UserProfile) (*entity.UserProfile, error) {
	if p == nil {
		return nil, nil
	}

	ent := &entity.UserProfile{}

	if p.UserID != "" {
		ent.UserID = p.UserID
	}

	// Strings
	if p.DisplayName != nil {
		ent.DisplayName = *p.DisplayName
	}

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

	if p.Birthdate.IsZero() {
		ent.Birthdate = null.TimeFrom(p.Birthdate)
	}

	// Scalars
	if p.HeightCM != 0 {
		ent.HeightCM = null.Int16From(p.HeightCM)
	}

	// FKs (SMALLINT → int16)
	if p.GenderID != 0 {
		ent.GenderID = null.Int16From(p.GenderID)
	}

	if p.DatingIntentionID != 0 {
		ent.DatingIntentionID = null.Int16From(p.DatingIntentionID)
	}

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

	if p.ChildrenStatusID != 0 {
		ent.ChildrenStatusID = null.Int16From(p.ChildrenStatusID)
	}

	if p.FamilyPlanID != nil {
		ent.FamilyPlanID = null.Int16From(int16(*p.FamilyPlanID))
	}

	if p.EthnicityID != 0 {
		ent.EthnicityID = null.Int16From(p.EthnicityID)
	}

	if p.CoverPhotoURL != nil {
		ent.CoverPhotoURL = null.StringFromPtr(p.CoverPhotoURL)
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

// ---- PREFERENCES ----

func MapPreferencesToEntity(userID string, pr *domain.Preferences) (*entity.UserPreference, error) {
	if pr == nil {
		return nil, nil
	}

	ent := &entity.UserPreference{
		UserID: userID,
	}

	if pr.DistanceKM != nil {
		ent.DistanceKM = null.Int16From(*pr.DistanceKM)
	}

	if pr.AgeMin != nil {
		ent.AgeMin = null.Int16From(*pr.AgeMin)
	}

	if pr.AgeMax != nil {
		ent.AgeMax = null.Int16From(*pr.AgeMax)
	}

	// Arrays: entity uses types.Int64Array
	if pr.SeekGenderIDs != nil {
		ent.SeekGenderIds = ids32ToI64Array(*pr.SeekGenderIDs)
	}

	if pr.SeekIntentionIDs != nil {
		ent.SeekIntentionIds = ids32ToI64Array(*pr.SeekIntentionIDs)
	}

	if pr.SeekReligionIDs != nil {
		ent.SeekReligionIds = ids32ToI64Array(*pr.SeekReligionIDs)
	}

	if pr.SeekPoliticalIDs != nil {
		ent.SeekPoliticalBeliefIds = ids32ToI64Array(*pr.SeekPoliticalIDs)
	}

	return ent, nil
}

// Helper: []int32 → types.Int64Array
func ids32ToI64Array(in []int32) types.Int64Array {
	if len(in) == 0 {
		return types.Int64Array{}
	}

	out := make(types.Int64Array, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}

	return out
}
