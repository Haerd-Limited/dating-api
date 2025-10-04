package dto

type ProfileCard struct {
	DisplayName string `json:"display_name"`
	Birthdate   string `json:"birthdate"`
	Age         int    `json:"age"`
	HeightCM    int16  `json:"height_cm"`
	UserID      string `json:"user_id"`
	Emoji       string `json:"emoji"`

	// Location
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"home_town"`
	Country   string  `json:"country"`

	Gender          string        `json:"gender"`
	DatingIntention string        `json:"dating_intention"`
	Religion        string        `json:"religion"`
	EducationLevel  string        `json:"education_level"`
	PoliticalBelief string        `json:"political_belief"`
	Drinking        string        `json:"drinking"`
	Smoking         string        `json:"smoking"`
	Marijuana       string        `json:"marijuana"`
	Drugs           string        `json:"drugs"`
	ChildrenStatus  *string       `json:"children_status"`
	FamilyPlan      *string       `json:"family_plan"`
	Ethnicity       string        `json:"ethnicity"`
	SpokenLanguages []string      `json:"spoken_languages"`
	VoicePrompts    []VoicePrompt `json:"voice_prompts"`
	Theme           UserTheme     `json:"theme"`
	CoverPhotoURL   string        `json:"cover_photo_url"`

	Verified bool `json:"verified"`

	Work       *string `json:"work"`
	JobTitle   *string `json:"job_title"`
	University *string `json:"university"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type UserTheme struct {
	BaseHex string   `json:"base_hex"`
	Palette []string `json:"palette"`
}

type VoicePrompt struct {
	ID            int64  `json:"id"`
	URL           string `json:"url"`
	PromptType    Prompt `json:"prompt_type"`
	IsPrimary     bool   `json:"is_primary"`
	Position      int16  `json:"position"`
	DurationMs    int    `json:"duration_ms"`
	CoverPhotoUrl string `json:"cover_photo_url"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
