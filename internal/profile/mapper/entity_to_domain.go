package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/geo"
)

// MapProfileToDomain maps an entity.UserProfile (DB model) to a domain.UserProfile.
func MapProfileToDomain(up *entity.UserProfile) *domain.Profile {
	if up == nil {
		return nil
	}

	d := &domain.Profile{
		UserID:         up.UserID,
		CreatedAt:      up.CreatedAt,
		UpdatedAt:      up.UpdatedAt,
		VerifiedStatus: up.Verified,
	}

	if up.Emoji.Valid {
		d.Emoji = up.Emoji.String
	}

	// Basic fields
	d.DisplayName = up.DisplayName

	if up.Birthdate.Valid {
		d.Birthdate = up.Birthdate.Time
	}

	if up.HeightCM.Valid {
		d.HeightCM = up.HeightCM.Int16
	}

	// Location
	if up.Geo != "" {
		d.Longitude, d.Latitude, _ = geo.ParseEWKBLonLatHex(up.Geo)
	}

	if up.City.Valid {
		d.City = up.City.String
	}

	if up.Country.Valid {
		d.Country = up.Country.String
	}

	// Single-select IDs
	if up.GenderID.Valid {
		d.GenderID = up.GenderID.Int16
	}

	if up.DatingIntentionID.Valid {
		d.DatingIntentionID = up.DatingIntentionID.Int16
	}

	if up.SexualityID.Valid {
		d.SexualityID = up.SexualityID.Int16
	}

	if up.ReligionID.Valid {
		d.ReligionID = up.ReligionID.Int16
	}

	if up.EducationLevelID.Valid {
		d.EducationLevelID = up.EducationLevelID.Int16
	}

	if up.PoliticalBeliefID.Valid {
		d.PoliticalBeliefID = up.PoliticalBeliefID.Int16
	}

	if up.DrinkingID.Valid {
		d.DrinkingID = up.DrinkingID.Int16
	}

	if up.SmokingID.Valid {
		d.SmokingID = up.SmokingID.Int16
	}

	if up.MarijuanaID.Valid {
		d.MarijuanaID = up.MarijuanaID.Int16
	}

	if up.DrugsID.Valid {
		d.DrugsID = up.DrugsID.Int16
	}

	if up.ChildrenStatusID.Valid {
		d.ChildrenStatusID = &up.ChildrenStatusID.Int16
	}

	if up.FamilyPlanID.Valid {
		d.FamilyPlanID = &up.FamilyPlanID.Int16
	}

	// EthnicityIDs will be loaded separately via repository

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

func MapLanguagesToDomain(g []*entity.Language) []domain.Language {
	if g == nil {
		return nil
	}

	var result []domain.Language

	for _, e := range g {
		result = append(result, domain.Language{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapEthnicityToDomain(g []*entity.Ethnicity) []domain.Ethnicity {
	if g == nil {
		return nil
	}

	var result []domain.Ethnicity

	for _, e := range g {
		result = append(result, domain.Ethnicity{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapEducationlevelsToDomain(g []*entity.EducationLevel) []domain.EducationLevel {
	if g == nil {
		return nil
	}

	var result []domain.EducationLevel

	for _, e := range g {
		result = append(result, domain.EducationLevel{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapPoliticalBeliefsToDomain(g []*entity.PoliticalBelief) []domain.PoliticalBelief {
	if g == nil {
		return nil
	}

	var result []domain.PoliticalBelief

	for _, e := range g {
		result = append(result, domain.PoliticalBelief{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapReligionsToDomain(g []*entity.Religion) []domain.Religion {
	if g == nil {
		return nil
	}

	var result []domain.Religion

	for _, e := range g {
		result = append(result, domain.Religion{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapHabitsToDomain(g []*entity.Habit) []domain.Habit {
	if g == nil {
		return nil
	}

	var result []domain.Habit

	for _, e := range g {
		result = append(result, domain.Habit{
			ID:    e.ID,
			Label: e.Label,
		})
	}

	return result
}

func MapPromptsToDomain(g []*entity.PromptType) []domain.Prompt {
	if g == nil {
		return nil
	}

	var result []domain.Prompt

	for _, e := range g {
		result = append(result, domain.Prompt{
			ID:       e.ID,
			Label:    e.Label,
			Key:      e.Key,
			Category: e.Category,
		})
	}

	return result
}

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
