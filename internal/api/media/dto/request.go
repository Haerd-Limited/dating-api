package dto

import "github.com/go-playground/validator/v10"

type TranscribeReelRequest struct {
	ReelURL string `json:"reel_url" validate:"required,url"`
}

func (r *TranscribeReelRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
