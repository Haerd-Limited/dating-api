package dto

import (
	"github.com/go-playground/validator/v10"
)

type ProfileRequest struct {
	ProfileBaseColour            string   `json:"profile_base_colour" validate:"required"`
	ProfileCoverMediaURL         string   `json:"profile_cover_media_url" validate:"required"`
	ProfileCoverMediaType        *string  `json:"profile_cover_media_type,omitempty"`
	ProfileCoverMediaAspectRatio *float64 `json:"profile_cover_media_aspect_ratio,omitempty"`
}

func (pr ProfileRequest) Validate() error {
	return validator.New().Struct(pr)
}

type VideoVerificationRequest struct {
	VideoKey string `json:"video_key" validate:"required"`
}

func (vvr VideoVerificationRequest) Validate() error {
	return validator.New().Struct(vvr)
}

type PromptsRequest struct {
	UploadedPrompts []VoicePrompt `json:"uploaded_prompts" validate:"required,min=4,max=6,dive"`
}

type VoicePrompt struct {
	URL                   string   `json:"upload_url" validate:"required"`
	PromptType            int16    `json:"prompt_type" validate:"required"`
	IsPrimary             bool     `json:"is_primary"`
	Position              int16    `json:"position" validate:"required"`
	CoverMediaURL         *string  `json:"cover_media_url,omitempty"`
	CoverMediaType        *string  `json:"cover_media_type,omitempty"`
	CoverMediaAspectRatio *float64 `json:"cover_media_aspect_ratio,omitempty"`
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
	EducationLevelID int16   `json:"education_level_id" validate:"required"`
	EthnicityIDs     []int16 `json:"ethnicity_ids" validate:"required,min=1,dive,gt=0"`
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
	SexualityID       int16  `json:"sexuality_id" validate:"required"`
}

func (br BasicsRequest) Validate() error {
	return validator.New().Struct(br)
}

type IntroRequest struct {
	Email string `json:"email" validate:"required,email"`
	// PhoneNumber string  `json:"phone_number" validate:"required"`
	FirstName            string  `json:"first_name" validate:"required"`
	LastName             *string `json:"last_name"`
	HowDidYouHearAboutUs *string `json:"how_did_you_hear_about_us"`
}

func (rr IntroRequest) Validate() error {
	return validator.New().Struct(rr)
}
