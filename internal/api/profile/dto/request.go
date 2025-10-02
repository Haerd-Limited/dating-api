package dto

import "github.com/go-playground/validator/v10"

type SwipesRequest struct {
	TargetUserID   string  `json:"target_user_id" validate:"required"`
	Action         string  `json:"action" validate:"oneof=like pass superlike"`
	IdempotencyKey *string `json:"idempotency_key"`
}

func (sr SwipesRequest) Validate() error {
	return validator.New().Struct(sr)
}

type UpdateProfileRequest struct {
	// Profile
	DisplayName   *string `json:"display_name,omitempty"`
	Birthdate     *string `json:"birthdate,omitempty"` // "YYYY-MM-DD"
	HeightCM      *int16  `json:"height_cm,omitempty"`
	ProfileEmoji  *string `json:"profile_emoji,omitempty"`
	CoverPhotoUrl *string `json:"cover_photo_url,omitempty"`

	// Location
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	City      *string  `json:"home_town,omitempty"`
	Country   *string  `json:"country,omitempty"`

	GenderID          *int16 `json:"gender_id,omitempty"`
	DatingIntentionID *int16 `json:"dating_intention_id,omitempty"`
	ReligionID        *int16 `json:"religion_id,omitempty"`
	EducationLevelID  *int16 `json:"education_level_id,omitempty"`
	PoliticalBeliefID *int16 `json:"political_belief_id,omitempty"`
	DrinkingID        *int16 `json:"drinking_id,omitempty"`
	SmokingID         *int16 `json:"smoking_id,omitempty"`
	MarijuanaID       *int16 `json:"marijuana_id,omitempty"`
	DrugsID           *int16 `json:"drugs_id,omitempty"`
	ChildrenStatusID  *int16 `json:"children_status_id,omitempty"`
	FamilyPlanID      *int16 `json:"family_plan_id,omitempty"`
	EthnicityID       *int16 `json:"ethnicity_id,omitempty"`

	Work            *string              `json:"work,omitempty"`
	JobTitle        *string              `json:"job_title,omitempty"`
	University      *string              `json:"university,omitempty"`
	SpokenLanguages []int16              `json:"spoken_languages,omitempty"`
	VoicePrompts    []VoicePromptRequest `json:"voice_prompts,omitempty"`
	Photos          []Photo              `json:"photos,omitempty"`
}

func (upr UpdateProfileRequest) Validate() error {
	return validator.New().Struct(upr)
}

type VoicePromptRequest struct {
	URL           string `json:"url"`
	PromptType    int16  `json:"prompt_type"`
	IsPrimary     bool   `json:"is_primary"`
	Position      int16  `json:"position"`
	DurationMs    int    `json:"duration_ms"`
	CoverPhotoURL string `json:"cover_photo_url"`
}
