package domain

type Profile struct {
	UserID               string
	ProfileBaseColour    string
	ProfileCoverPhotoURL string
}

type Prompts struct {
	UploadedPrompts []VoicePrompt
	UserID          string
}

type StepResult struct {
	AccessToken  *string
	RefreshToken *string
	OnboardingSteps
	Content any
}

type VoicePrompt struct {
	URL           string
	PromptType    int16
	IsPrimary     bool
	Position      int16
	DurationMs    int
	CoverPhotoUrl *string
}
type PhotosContent struct {
	Prompts                []Prompt
	VoicePromptsUploadUrls []UploadUrl
}

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}

type UploadedPhotos struct {
	UserID string
	Photos []Photo
}

type Photo struct {
	URL       string
	Position  int16
	IsPrimary bool
}

type Languages struct {
	UserID      string
	LanguageIDs []int16
}

type LanguagesContent struct {
	PhotoUploadUrls []UploadUrl
}

type UploadUrl struct {
	Key       string
	UploadUrl string
	Headers   map[string]string
	MaxBytes  int64
}

type WorkAndEducation struct {
	UserID     string
	Workplace  string
	JobTitle   string
	University string
}

type WorkAndEducationContent struct {
	Languages []Language
}

type Language struct {
	ID    int16
	Label string
}

type Background struct {
	UserID           string
	EducationLevelID int16
	EthnicityIDs     []int16
}

type Beliefs struct {
	UserID             string
	PoliticalBeliefsID int16
	ReligionID         int16
}

type BeliefsContent struct {
	EducationLevels []EducationLevel
	Ethnicities     []Ethnicity
}

type Ethnicity struct {
	ID    int16
	Label string
}

type EducationLevel struct {
	ID    int16
	Label string
}

type Lifestyle struct {
	UserID      string
	DrinkingID  int16
	SmokingID   int16
	MarijuanaID int16
	DrugsID     int16
}

type LifestyleContent struct {
	Religions        []Religion
	PoliticalBeliefs []PoliticalBelief
}

type Religion struct {
	ID    int16
	Label string
}

type PoliticalBelief struct {
	ID    int16
	Label string
}

type Location struct {
	UserID    string
	Latitude  float64
	Longitude float64
	City      string
	Country   string
}

type LocationContent struct {
	Habits []Habit
}

type Habit struct {
	ID    int16
	Label string
}

type Basics struct {
	UserID            string
	Birthdate         string
	HeightCm          int16
	GenderID          int16
	DatingIntentionID int16
}

type Intro struct {
	Email     string
	UserID    string
	FirstName string
	LastName  *string
}

type OnboardingSteps struct {
	PreviousStep Steps
	CurrentStep  Steps
	NextStep     Steps
	Progress     float64
	Steps        []Steps
	TotalSteps   int
}

type IntroContent struct {
	DatingIntentions []DatingIntention
	Genders          []Gender
}

type Gender struct {
	ID    int16
	Label string
}

type DatingIntention struct {
	ID    int16
	Label string
}

type Preferences struct {
	// Preferences
	DistanceKM       *int16
	AgeMin           *int16
	AgeMax           *int16
	SeekGenderIDs    *[]int32
	SeekIntentionIDs *[]int32
	SeekReligionIDs  *[]int32
	SeekPoliticalIDs *[]int32
}

// PreregistrationStats summarizes the current preregistration counts and limits.
type PreregistrationStats struct {
	MaleCount    int64 `json:"male_count"`
	FemaleCount  int64 `json:"female_count"`
	MaxTotal     int   `json:"max_total"`
	MaxMale      int   `json:"max_male"`
	MaxFemale    int   `json:"max_female"`
	CapEnforced  bool  `json:"cap_enforced"`
	TotalCurrent int64 `json:"total_current"`
}
