package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

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
		ent.DisplayName = null.StringFrom(*p.DisplayName)
	}

	if p.City != nil {
		ent.City = null.StringFrom(*p.City)
	}

	if p.Country != nil {
		ent.Country = null.StringFrom(*p.Country)
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

	if p.ReligionID != nil {
		ent.ReligionID = null.Int16From(int16(*p.ReligionID))
	}

	if p.EducationLevelID != nil {
		ent.EducationLevelID = null.Int16From(int16(*p.EducationLevelID))
	}

	if p.PoliticalBeliefID != nil {
		ent.PoliticalBeliefID = null.Int16From(int16(*p.PoliticalBeliefID))
	}

	if p.DrinkingID != nil {
		ent.DrinkingID = null.Int16From(int16(*p.DrinkingID))
	}

	if p.SmokingID != nil {
		ent.SmokingID = null.Int16From(int16(*p.SmokingID))
	}

	if p.MarijuanaID != nil {
		ent.MarijuanaID = null.Int16From(int16(*p.MarijuanaID))
	}

	if p.DrugsID != nil {
		ent.DrugsID = null.Int16From(int16(*p.DrugsID))
	}

	if p.ChildrenStatusID != nil {
		ent.ChildrenStatusID = null.Int16From(int16(*p.ChildrenStatusID))
	}

	if p.FamilyPlanID != nil {
		ent.FamilyPlanID = null.Int16From(int16(*p.FamilyPlanID))
	}

	if p.EthnicityID != nil {
		ent.EthnicityID = null.Int16From(int16(*p.EthnicityID))
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
	if p.Latitude != nil && p.Longitude != nil {
		ent.Geo = fmt.Sprintf("SRID=4326;POINT(%f %f)", *p.Longitude, *p.Latitude)
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
