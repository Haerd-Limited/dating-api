package dto

type AuthResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         *User  `json:"user,omitempty"`
}

type User struct {
	ID             string `json:"id"`
	OnboardingStep string `json:"onboarding_step"`
}

type RequestCodeResponse struct {
	// Return generic OK to avoid user enumeration, plus optional masking hint
	SentTo         string  `json:"sent_to"`         // e.g., "e***@example.com" or "+44******123"
	OnboardingStep *string `json:"onboarding_step"` // Only set when user exists during registration and is in pre-registration
}
