package dto

import (
	"github.com/go-playground/validator/v10"
)

type UpdateOnboardingRequest struct {
	// Profile
	DisplayName *string `json:"display_name,omitempty"`
	Birthdate   *string `json:"birthdate,omitempty"` // "YYYY-MM-DD"
	HeightCM    *int16  `json:"height_cm,omitempty"`

	// Location
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	City      *string  `json:"city,omitempty"`
	Country   *string  `json:"country,omitempty"`

	// Single-selects (FKs)
	GenderID          *int32 `json:"gender_id,omitempty"`
	DatingIntentionID *int32 `json:"dating_intention_id,omitempty"`
	ReligionID        *int32 `json:"religion_id,omitempty"`
	EducationLevelID  *int32 `json:"education_level_id,omitempty"`
	PoliticalBeliefID *int32 `json:"political_belief_id,omitempty"`
	DrinkingID        *int32 `json:"drinking_id,omitempty"`
	SmokingID         *int32 `json:"smoking_id,omitempty"`
	MarijuanaID       *int32 `json:"marijuana_id,omitempty"`
	DrugsID           *int32 `json:"drugs_id,omitempty"`
	ChildrenStatusID  *int32 `json:"children_status_id,omitempty"`
	FamilyPlanID      *int32 `json:"family_plan_id,omitempty"`
	EthnicityID       *int32 `json:"ethnicity_id,omitempty"`

	// Preferences
	DistanceKM       *int16   `json:"distance_km,omitempty"`
	AgeMin           *int16   `json:"age_min,omitempty"`
	AgeMax           *int16   `json:"age_max,omitempty"`
	SeekGenderIDs    *[]int32 `json:"seek_gender_ids,omitempty"`
	SeekIntentionIDs *[]int32 `json:"seek_intention_ids,omitempty"`
	SeekReligionIDs  *[]int32 `json:"seek_religion_ids,omitempty"`
	SeekPoliticalIDs *[]int32 `json:"seek_political_belief_ids,omitempty"`

	// Extra text fields in user_profiles
	Work        *string         `json:"work,omitempty"`
	JobTitle    *string         `json:"job_title,omitempty"`
	University  *string         `json:"university,omitempty"`
	ProfileMeta *map[string]any `json:"profile_meta,omitempty"`

	// Multi-selects (FULL REPLACE if provided)
	LanguageIDs *[]int32 `json:"language_ids,omitempty"`
	InterestIDs *[]int32 `json:"interest_ids,omitempty"`

	// Optional: bump progress
	BumpOnboardingStep *bool `json:"bump_onboarding_step,omitempty"`
}

func (uor UpdateOnboardingRequest) Validate() error {
	return validator.New().Struct(uor)
}
