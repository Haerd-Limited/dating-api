package dto

import (
	"errors"
	"strings"
)

type DeviceTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

func (r *DeviceTokenRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return errors.New("token is required")
	}

	return nil
}
