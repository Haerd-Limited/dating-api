package dto

import (
	"github.com/go-playground/validator/v10"
)

type PromptsRequest struct {
	UploadedPrompts []VoicePrompt `json:"uploaded_prompts" validate:"required,dive"`
}

type VoicePrompt struct {
	URL        string `json:"upload_url" validate:"required"`
	PromptType int16  `json:"prompt_type" validate:"required"`
	IsPrimary  bool   `json:"is_primary"`
	Position   int16  `json:"position" validate:"required"`
}

func (pr PromptsRequest) Validate() error {
	return validator.New().Struct(pr)
}

type PhotosRequest struct {
	UploadedPhotos []Photo `json:"uploaded_photos" validate:"required,dive"`
}

type Photo struct {
	URL       string `json:"upload_url" validate:"required"`
	Position  int16  `json:"position" validate:"required"`
	IsPrimary bool   `json:"is_primary"`
}

func (pr PhotosRequest) Validate() error {
	return validator.New().Struct(pr)
}

type LanguagesRequest struct {
	LanguageIDs []int16 `json:"language_ids" validate:"required,dive,gt=0"` // todo: find out what this means
}

func (wae LanguagesRequest) Validate() error {
	return validator.New().Struct(wae)
}

type WorkAndEducationRequest struct {
	Workplace  string `json:"workplace"`
	JobTitle   string `json:"job_title"`
	University string `json:"university"`
}

func (wae WorkAndEducationRequest) Validate() error {
	return validator.New().Struct(wae)
}

type BackgroundRequest struct {
	EducationLevelID int16 `json:"education_level_id" validate:"required"`
	EthnicityID      int16 `json:"ethnicity_id" validate:"required"`
}

func (br BackgroundRequest) Validate() error {
	return validator.New().Struct(br)
}

type BeliefsRequest struct {
	PoliticalBeliefID int16 `json:"political_belief_id" validate:"required"`
	ReligionID        int16 `json:"religion_id" validate:"required"`
}

func (br BeliefsRequest) Validate() error {
	return validator.New().Struct(br)
}

type LifestyleRequest struct {
	DrinkingID  int16 `json:"drinking_id" validate:"required"`
	SmokingID   int16 `json:"smoking_id" validate:"required"`
	MarijuanaID int16 `json:"marijuana_id" validate:"required"`
	DrugsID     int16 `json:"drugs_id" validate:"required"`
}

func (lr LifestyleRequest) Validate() error {
	return validator.New().Struct(lr)
}

type LocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	City      string  `json:"home_town" validate:"required"`
	Country   string  `json:"country" validate:"required"`
}

func (lr LocationRequest) Validate() error {
	return validator.New().Struct(lr)
}

type BasicsRequest struct {
	Birthdate         string `json:"birthdate" validate:"required"`
	HeightCm          int16  `json:"height_cm" validate:"required"`
	GenderID          int16  `json:"gender_id" validate:"required"`
	DatingIntentionID int16  `json:"dating_intention_id" validate:"required"`
}

func (br BasicsRequest) Validate() error {
	return validator.New().Struct(br)
}

type RegisterRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phone_number" validate:"required"`
	FirstName   string  `json:"first_name" validate:"required"`
	LastName    *string `json:"last_name"`
}

func (rr RegisterRequest) Validate() error {
	return validator.New().Struct(rr)
}

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
