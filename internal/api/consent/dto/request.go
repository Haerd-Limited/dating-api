package dto

import "github.com/go-playground/validator/v10"

type RecordConsentRequest struct {
	Type     string `json:"type" validate:"required,oneof=privacy_policy terms_of_service"`
	Version  string `json:"version" validate:"required"`
	Accepted bool   `json:"accepted"`
}

func (r RecordConsentRequest) Validate() error {
	return validator.New().Struct(r)
}

type ConsentResponse struct {
	Type       string  `json:"type"`
	Version    string  `json:"version"`
	Accepted   bool    `json:"accepted"`
	AcceptedAt string  `json:"accepted_at"`
	RevokedAt  *string `json:"revoked_at,omitempty"`
}

type ListConsentsResponse struct {
	Consents []ConsentResponse `json:"consents"`
}
