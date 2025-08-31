package dto

import (
	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

type RequestCodeRequest struct {
	Channel string  `json:"channel" validate:"required,oneof=email sms"`
	Email   *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone   *string `json:"phone,omitempty"`   // validate E.164 in service
	Purpose string  `json:"purpose,omitempty"` // default "login"
}

func (rcr RequestCodeRequest) Validate() error {
	return validator.New().Struct(rcr)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (lr LoginRequest) Validate() error {
	return validator.New().Struct(lr)
}

func (rfr RefreshRequest) Validate() error {
	return validator.New().Struct(rfr)
}

func (lor LogoutRequest) Validate() error {
	return validator.New().Struct(lor)
}
