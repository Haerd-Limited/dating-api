package dto

import (
	"errors"
	"strings"
	"time"
)

type InitiateRevealResponse struct {
	Message string `json:"message"`
}

type ConfirmRevealResponse struct {
	Photos []Photo `json:"photos"`
}

type MakeRevealDecisionRequest struct {
	Decision string `json:"decision" validate:"required,oneof=continue date unmatch"`
}

func (r *MakeRevealDecisionRequest) Validate() error {
	if r.Decision == "" {
		return errors.New("decision is required")
	}

	validDecisions := []string{"continue", "date", "unmatch"}
	for _, valid := range validDecisions {
		if r.Decision == valid {
			return nil
		}
	}

	return errors.New("decision must be one of: continue, date, unmatch")
}

type MakeRevealDecisionResponse struct {
	Message string `json:"message"`
}

type GetMatchPhotosResponse struct {
	Photos []Photo `json:"photos"`
}

type Photo struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	Position  int16  `json:"position"`
}

type RevealRequest struct {
	ConversationID string    `json:"conversation_id"`
	InitiatorID    string    `json:"initiator_id"`
	RequestedAt    time.Time `json:"requested_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Status         string    `json:"status"`
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
