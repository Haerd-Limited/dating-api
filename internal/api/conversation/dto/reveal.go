package dto

import (
	"errors"
	"strings"
)

type Photo struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	Position  int16  `json:"position"`
}

type UnmatchRequest struct {
	Reason string `json:"reason" validate:"required"`
}

func (r *UnmatchRequest) Validate() error {
	if r.Reason == "" {
		return errors.New("reason is required")
	}

	// Trim whitespace and check if empty after trimming
	trimmed := strings.TrimSpace(r.Reason)
	if trimmed == "" {
		return errors.New("reason cannot be empty or whitespace only")
	}

	return nil
}

type UnmatchResponse struct {
	Message string `json:"message"`
}
