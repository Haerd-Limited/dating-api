package dto

type WaitlistUserResponse struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	Phone          string `json:"phone"`
	OnboardingStep string `json:"onboarding_step"`
	CreatedAt      string `json:"created_at"`
	Contacted      bool   `json:"contacted"`
}

type RecipientResultResponse struct {
	UserID string  `json:"user_id"`
	Phone  string  `json:"phone"`
	Status string  `json:"status"`
	Error  *string `json:"error,omitempty"`
}

type BroadcastResultResponse struct {
	Total      int                       `json:"total"`
	Sent       int                       `json:"sent"`
	Failed     int                       `json:"failed"`
	Recipients []RecipientResultResponse `json:"recipients"`
}
