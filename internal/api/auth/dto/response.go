package dto

type AuthResponse struct { // TOdo: uniffy with register response and mapper
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID       string `json:"user_id,omitempty"`
}

type RequestCodeResponse struct {
	// Return generic OK to avoid user enumeration, plus optional masking hint
	SentTo string `json:"sent_to"` // e.g., "e***@example.com" or "+44******123"
}
