package dto

import "github.com/go-playground/validator/v10"

type CompleteRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

func (cr CompleteRequest) Validate() error {
	return validator.New().Struct(cr)
}
