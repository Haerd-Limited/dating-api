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

type PhotosContent struct {
	Prompts                []Prompt    `json:"prompts"`
	VoicePromptsUploadUrls []UploadUrl `json:"prompt_upload_urls"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

type LanguagesContent struct {
	PhotoUploadUrls []UploadUrl `json:"photo_upload_urls"`
}

type UploadUrl struct {
	Key       string            `json:"key"`
	UploadUrl string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
	MaxBytes  int64             `json:"max_bytes"`
}

type WorkAndEducationContent struct {
	Languages []Language `json:"languages"`
}

type Language struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type BeliefsContent struct {
	EducationLevels []EducationLevel `json:"education_options"`
	Ethnicities     []Ethnicity      `json:"ethnicity_options"`
}

type Ethnicity struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type EducationLevel struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type LifestyleContent struct {
	Religions        []Religion        `json:"religion_options"`
	PoliticalBeliefs []PoliticalBelief `json:"political_belief_options"`
}

type Religion struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type PoliticalBelief struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type LocationContent struct {
	Habits []Habit `json:"habit_options"`
}

type Habit struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type IntroContent struct {
	Genders     []Gender    `json:"genders"`
	Sexualities []Sexuality `json:"sexualities"`
}

type Gender struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Sexuality struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}
