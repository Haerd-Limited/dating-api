package dto

type GetProfileResponse struct {
	Profile Profile `json:"profile"`
}

type Profile struct {
	DisplayName *string `json:"display_name,omitempty"`
	Birthdate   string  `json:"birthdate"`
	Age         int     `json:"age"`
	HeightCM    int16   `json:"height_cm"`
	UserID      string  `json:"user_id"`

	// Location
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	City      string  `json:"home_town,omitempty"`
	Country   string  `json:"country,omitempty"`

	Gender          Gender          `json:"gender,omitempty"`
	DatingIntention DatingIntention `json:"dating_intention"`
	Religion        Religion        `json:"religion"`
	EducationLevel  EducationLevel  `json:"education_level"`
	PoliticalBelief PoliticalBelief `json:"political_belief"`
	Drinking        Habit           `json:"drinking"`
	Smoking         Habit           `json:"smoking"`
	Marijuana       Habit           `json:"marijuana"`
	Drugs           Habit           `json:"drugs"`
	ChildrenStatus  *Status         `json:"children_status,omitempty"`
	FamilyPlan      *Status         `json:"family_plan,omitempty"`
	Ethnicity       Ethnicity       `json:"ethnicity"`
	SpokenLanguages []Language      `json:"spoken_languages"`
	VoicePrompts    []VoicePrompt   `json:"voice_prompts"`

	Work       *string `json:"work,omitempty"`
	JobTitle   *string `json:"job_title,omitempty"`
	University *string `json:"university,omitempty"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type VoicePrompt struct {
	URL        string `json:"url"`
	PromptType Prompt `json:"prompt_type"`
	IsPrimary  bool   `json:"is_primary"`
	Position   int16  `json:"position"`
	DurationMs int    `json:"duration_ms"`
}

type Status struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

type Language struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Ethnicity struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type EducationLevel struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Religion struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type PoliticalBelief struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Gender struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type DatingIntention struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Habit struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}
