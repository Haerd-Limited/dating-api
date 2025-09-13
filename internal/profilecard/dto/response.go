package dto

type ProfileCard struct {
	DisplayName string `json:"display_name"`
	Birthdate   string `json:"birthdate"`
	Age         int    `json:"age"`
	HeightCM    int16  `json:"height_cm"`
	UserID      string `json:"user_id"`

	// Location
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	City      string  `json:"home_town,omitempty"`
	Country   string  `json:"country,omitempty"`

	Gender          string        `json:"gender,omitempty"`
	DatingIntention string        `json:"dating_intention"`
	Religion        string        `json:"religion"`
	EducationLevel  string        `json:"education_level"`
	PoliticalBelief string        `json:"political_belief"`
	Drinking        string        `json:"drinking"`
	Smoking         string        `json:"smoking"`
	Marijuana       string        `json:"marijuana"`
	Drugs           string        `json:"drugs"`
	ChildrenStatus  *string       `json:"children_status,omitempty"`
	FamilyPlan      *string       `json:"family_plan,omitempty"`
	Ethnicity       string        `json:"ethnicity"`
	SpokenLanguages []string      `json:"spoken_languages"`
	VoicePrompts    []VoicePrompt `json:"voice_prompts"`
	Theme           UserTheme     `json:"theme"`

	Verified bool `json:"verified"`

	Work       *string `json:"work,omitempty"`
	JobTitle   *string `json:"job_title,omitempty"`
	University *string `json:"university,omitempty"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type UserTheme struct {
	BaseHex string   `json:"base_hex"`
	Palette []string `json:"palette"`
}

type VoicePrompt struct {
	URL        string `json:"url"`
	PromptType Prompt `json:"prompt_type"`
	IsPrimary  bool   `json:"is_primary"`
	Position   int16  `json:"position"`
	DurationMs int    `json:"duration_ms"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
