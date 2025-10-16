package dto

// Result is what handlers return to clients
type StartResponse struct {
	SessionID string `json:"session_id"`
	Region    string `json:"region"`
}

type CompleteResponse struct {
	Status        string   `json:"status"` // passed / failed / needs_review
	MatchScore    float64  `json:"match_score,omitempty"`
	Reasons       []string `json:"reasons,omitempty"`
	PhotoVerified bool     `json:"photo_verified"`
}
