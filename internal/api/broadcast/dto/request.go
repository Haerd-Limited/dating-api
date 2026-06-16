package dto

import "github.com/go-playground/validator/v10"

type SendBroadcastRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,dive,required"`
	Message string   `json:"message" validate:"required"`
}

func (r SendBroadcastRequest) Validate() error {
	return validator.New().Struct(r)
}
