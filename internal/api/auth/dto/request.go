package dto

import (
	"github.com/go-playground/validator/v10"
)

type RegisterRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"PhoneNumber" validate:"required"`
	FirstName   string  `json:"firstName" validate:"required"`
	LastName    *string `json:"lastName"`
	DateOfBirth string  `json:"dateOfBirth" validate:"required"`
}
type LoginRequest struct {
	PhoneNumber string `json:"PhoneNumber" validate:"required"`
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

func (rr RegisterRequest) Validate() error {
	return validator.New().Struct(rr)
}
