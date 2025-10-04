package dto

type GetProfileResponse struct {
	Profile Profile `json:"profile"`
}

type Profile struct {
	DisplayName   string `json:"display_name"`
	Birthdate     string `json:"birthdate"`
	Age           int    `json:"age"`
	HeightCM      int16  `json:"height_cm"`
	UserID        string `json:"user_id"`
	CoverPhotoURL string `json:"cover_photo_url"`
	Emoji         string `json:"emoji"`

	// Location
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"home_town"`
	Country   string  `json:"country"`

	Gender          Gender          `json:"gender"`
	DatingIntention DatingIntention `json:"dating_intention"`
	Religion        Religion        `json:"religion"`
	EducationLevel  EducationLevel  `json:"education_level"`
	PoliticalBelief PoliticalBelief `json:"political_belief"`
	Drinking        Habit           `json:"drinking"`
	Smoking         Habit           `json:"smoking"`
	Marijuana       Habit           `json:"marijuana"`
	Drugs           Habit           `json:"drugs"`
	ChildrenStatus  *Status         `json:"children_status"`
	FamilyPlan      *Status         `json:"family_plan"`
	Ethnicity       Ethnicity       `json:"ethnicity"`
	SpokenLanguages []Language      `json:"spoken_languages"`
	VoicePrompts    []VoicePrompt   `json:"voice_prompts"`
	Photos          []Photo         `json:"photos"`

	Work       *string   `json:"work"`
	JobTitle   *string   `json:"job_title"`
	University *string   `json:"university"`
	Theme      UserTheme `json:"theme"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserTheme struct {
	BaseHex string   `json:"base_hex"`
	Palette []string `json:"palette"`
}

type Photo struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	Position  int16  `json:"position"`
}

type VoicePrompt struct {
	URL           string `json:"url"`
	PromptType    Prompt `json:"prompt_type"`
	IsPrimary     bool   `json:"is_primary"`
	Position      int16  `json:"position"`
	DurationMs    int    `json:"duration_ms"`
	CoverPhotoURL string `json:"cover_photo_url"`
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
