package dto

type OnboardingResponse struct {
	AccessToken  *string `json:"access_token,omitempty"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	OnboardingSteps
	Content any `json:"content"`
}

type OnboardingSteps struct {
	PreviousStep string   `json:"previous_step"`
	CurrentStep  string   `json:"current_step"`
	NextStep     string   `json:"next_step"`
	Progress     float64  `json:"progress"`
	Steps        []string `json:"steps"`
	TotalSteps   int      `json:"total_steps"`
}

type RegisterContent struct {
	DatingIntentions []DatingIntention `json:"dating_intentions"`
	Genders          []Gender          `json:"genders"`
}

type Gender struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type DatingIntention struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}
