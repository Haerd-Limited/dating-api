package dto

import "github.com/go-playground/validator/v10"

type GenerateVoiceNoteUploadUrlRequest struct {
	Purpose string `json:"purpose" validate:"required,oneof=voicenote prompt"`
}

func (gvn GenerateVoiceNoteUploadUrlRequest) Validate() error {
	return validator.New().Struct(gvn)
}
