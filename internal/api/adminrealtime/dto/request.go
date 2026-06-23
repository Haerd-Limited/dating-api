package dto

import "github.com/go-playground/validator/v10"

type PresenceRequest struct {
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id" validate:"required"`
}

func (r PresenceRequest) Validate() error {
	return validator.New().Struct(r)
}
