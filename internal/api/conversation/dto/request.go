package dto

import "github.com/go-playground/validator/v10"

type SendMessageRequest struct {
	ClientMsgID  string   `json:"client_msg_id" validate:"required"`
	Type         string   `json:"type" validate:"required,oneof=text voice"`
	TextBody     *string  `json:"text_body"`
	MediaKey     *string  `json:"media_key"`
	MediaSeconds *float64 `json:"media_seconds"`
}

func (smr SendMessageRequest) Validate() error {
	return validator.New().Struct(smr)
}
