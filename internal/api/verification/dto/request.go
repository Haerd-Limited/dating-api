package dto

import "github.com/go-playground/validator/v10"

type CompleteRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

func (cr CompleteRequest) Validate() error {
	return validator.New().Struct(cr)
}

type SubmitVideoRequest struct {
	VideoKey string `json:"video_key" validate:"required"`
}

func (svr SubmitVideoRequest) Validate() error {
	return validator.New().Struct(svr)
}

type ApproveVideoRequest struct {
	Notes *string `json:"notes,omitempty"`
}

func (avr ApproveVideoRequest) Validate() error {
	return validator.New().Struct(avr)
}

type RejectVideoRequest struct {
	RejectionReason string  `json:"rejection_reason" validate:"required"`
	Notes           *string `json:"notes,omitempty"`
}

func (rvr RejectVideoRequest) Validate() error {
	return validator.New().Struct(rvr)
}
