package dto

import "github.com/go-playground/validator/v10"

type CreateSessionRequest struct {
	DisplayName string `json:"display_name" validate:"required"`
}

func (r CreateSessionRequest) Validate() error {
	return validator.New().Struct(r)
}

type CreateSessionResponse struct {
	SessionToken string `json:"session_token"`
	DisplayName  string `json:"display_name"`
	ExpiresAt    string `json:"expires_at"`
}

type RosterResponse struct {
	Names []string `json:"names"`
}
