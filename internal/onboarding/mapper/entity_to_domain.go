package mapper

import (
	"fmt"
	"strings"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func MapGendersToDomain(g []*entity.Gender) []domain.Gender {
	if g == nil {
		return nil
	}

	var result []domain.Gender

	for _, e := range g {
		result = append(result, domain.Gender{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapDatingIntentionsToDomain(di []*entity.DatingIntention) []domain.DatingIntention {
	if di == nil {
		return nil
	}

	var result []domain.DatingIntention
	for _, e := range di {
		result = append(result, domain.DatingIntention{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

// MapUserProfileToDomain maps an entity.UserProfile (DB model) to a domain.UserProfile.
func MapUserProfileToDomain(up *entity.UserProfile) *domain.UserProfile {
	if up == nil {
		return nil
	}

	d := &domain.UserProfile{
		UserID:    up.UserID,
		CreatedAt: up.CreatedAt,
		UpdatedAt: up.UpdatedAt,
	}

	// Basic fields
	if up.DisplayName.Valid {
		d.DisplayName = &up.DisplayName.String
	}

	if up.Birthdate.Valid {
		d.Birthdate = up.Birthdate.Time
	}

	if up.HeightCM.Valid {
		d.HeightCM = up.HeightCM.Int16
	}

	// Location
	if up.Geo != "" {
		// Parse WKT-like value: "SRID=4326;POINT(lon lat)"
		// Example: "SRID=4326;POINT(-73.9857 40.7484)"
		parts := strings.Split(up.Geo, ";POINT(")
		if len(parts) == 2 {
			point := strings.TrimSuffix(parts[1], ")")

			var lon, lat float64

			if _, err := fmt.Sscanf(point, "%f %f", &lon, &lat); err == nil {
				d.Latitude = &lat
				d.Longitude = &lon
			}
		}
	}

	if up.City.Valid {
		d.City = &up.City.String
	}

	if up.Country.Valid {
		d.Country = &up.Country.String
	}

	// Single-select IDs
	if up.GenderID.Valid {
		d.GenderID = up.GenderID.Int16
	}

	if up.DatingIntentionID.Valid {
		d.DatingIntentionID = up.DatingIntentionID.Int16
	}

	if up.ReligionID.Valid {
		v := int32(up.ReligionID.Int16)
		d.ReligionID = &v
	}

	if up.EducationLevelID.Valid {
		v := int32(up.EducationLevelID.Int16)
		d.EducationLevelID = &v
	}

	if up.PoliticalBeliefID.Valid {
		v := int32(up.PoliticalBeliefID.Int16)
		d.PoliticalBeliefID = &v
	}

	if up.DrinkingID.Valid {
		v := int32(up.DrinkingID.Int16)
		d.DrinkingID = &v
	}

	if up.SmokingID.Valid {
		v := int32(up.SmokingID.Int16)
		d.SmokingID = &v
	}

	if up.MarijuanaID.Valid {
		v := int32(up.MarijuanaID.Int16)
		d.MarijuanaID = &v
	}

	if up.DrugsID.Valid {
		v := int32(up.DrugsID.Int16)
		d.DrugsID = &v
	}

	if up.ChildrenStatusID.Valid {
		v := int32(up.ChildrenStatusID.Int16)
		d.ChildrenStatusID = &v
	}

	if up.FamilyPlanID.Valid {
		v := int32(up.FamilyPlanID.Int16)
		d.FamilyPlanID = &v
	}

	if up.EthnicityID.Valid {
		v := int32(up.EthnicityID.Int16)
		d.EthnicityID = &v
	}

	// Extra text fields
	if up.Work.Valid {
		d.Work = &up.Work.String
	}

	if up.JobTitle.Valid {
		d.JobTitle = &up.JobTitle.String
	}

	if up.University.Valid {
		d.University = &up.University.String
	}

	// ProfileMeta (jsonb)
	if up.ProfileMeta.Valid {
		var meta map[string]any
		if err := up.ProfileMeta.Unmarshal(&meta); err == nil {
			d.ProfileMeta = &meta
		}
	}

	return d
}
