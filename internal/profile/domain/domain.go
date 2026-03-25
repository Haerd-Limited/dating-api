package domain

import "time"

const (
	VerificationStatusVerified    = "VERIFIED"
	VerificationStatusUnverified  = "UNVERIFIED"
	VerificationStatusUnderReview = "UNDER_REVIEW"
)

type Profile struct {
	DisplayName    string
	Birthdate      time.Time
	HeightCM       int16
	UserID         string
	VerifiedStatus string

	// Location
	Latitude  float64
	Longitude float64
	City      string
	Country   string

	// Single-selects
	GenderID              int16
	DatingIntentionID     int16
	SexualityID           int16
	ReligionID            int16
	EducationLevelID      int16
	PoliticalBeliefID     int16
	DrinkingID            int16
	SmokingID             int16
	MarijuanaID           int16
	DrugsID               int16
	ChildrenStatusID      *int16
	FamilyPlanID          *int16
	EthnicityIDs          []int16
	CoverMediaURL         *string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64
	Emoji                 string

	// Extra text fields in user_profiles

	// Work the user's workplace
	Work        *string
	JobTitle    *string
	University  *string
	ProfileMeta *map[string]any

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateProfile struct {
	DisplayName  *string
	Birthdate    *time.Time
	HeightCM     *int16
	UserID       string
	ProfileEmoji *string
	BaseColour   *string

	// Location
	Latitude  *float64
	Longitude *float64
	City      *string
	Country   *string

	// Single-selects
	GenderID              *int16
	DatingIntentionID     *int16
	SexualityID           *int16
	ReligionID            *int16
	EducationLevelID      *int16
	PoliticalBeliefID     *int16
	DrinkingID            *int16
	SmokingID             *int16
	MarijuanaID           *int16
	DrugsID               *int16
	ChildrenStatusID      *int16
	FamilyPlanID          *int16
	EthnicityIDs          []int16
	SpokenLanguages       []int16
	VoicePrompts          []VoicePromptUpdate
	Photos                []Photo
	CoverMediaURL         *string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64

	Work       *string
	JobTitle   *string
	University *string

	CreatedAt *time.Time
	UpdatedAt time.Time
}

type EnrichedProfile struct {
	DisplayName           string
	Birthdate             time.Time
	Age                   int
	HeightCM              int16
	UserID                string
	CoverMediaURL         *string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64
	Emoji                 string

	// Location
	Latitude  float64
	Longitude float64
	City      string
	Country   string

	Gender          Gender
	DatingIntention DatingIntention
	Sexuality       Sexuality
	Religion        Religion
	EducationLevel  EducationLevel
	PoliticalBelief PoliticalBelief
	Drinking        Habit
	Smoking         Habit
	Marijuana       Habit
	Drugs           Habit
	ChildrenStatus  *Status
	FamilyPlan      *Status
	Ethnicities     []Ethnicity
	SpokenLanguages []Language
	VoicePrompts    []ProfileVoicePrompt
	Photos          []Photo
	Theme           UserTheme
	Work            *string
	JobTitle        *string
	University      *string
	VerifiedStatus  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserTheme struct {
	BaseHex string
	Palette []string
}

type VoicePrompt struct {
	PromptID              int64
	Prompt                string
	VoiceNoteURL          string
	WaveformData          []float32
	CoverMediaURL         string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64
}

type ProfileVoicePrompt struct {
	ID                          int64
	URL                         string
	PromptType                  Prompt
	IsPrimary                   bool
	Position                    int16
	DurationMs                  int
	WaveformData                []float32
	PromptCoverMediaURL         string
	PromptCoverMediaType        *string
	PromptCoverMediaAspectRatio *float64
}
type VoicePromptUpdate struct {
	URL                         string
	PromptTypeID                int16
	IsPrimary                   bool
	Position                    int16
	DurationMs                  int
	WaveformData                []float32
	PromptCoverMediaURL         string
	PromptCoverMediaType        *string
	PromptCoverMediaAspectRatio *float64
}

type Photo struct {
	URL       string
	IsPrimary bool
	Position  int16
}

type Status struct {
	ID    int16
	Label string
	Key   string
}

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}

type Language struct {
	ID    int16
	Label string
}

type Ethnicity struct {
	ID    int16
	Label string
}

type EducationLevel struct {
	ID    int16
	Label string
}

type Religion struct {
	ID    int16
	Label string
}

type PoliticalBelief struct {
	ID    int16
	Label string
}

type Gender struct {
	ID    int16
	Label string
}

type DatingIntention struct {
	ID    int16
	Label string
}

type Habit struct {
	ID    int16
	Label string
}

type Sexuality struct {
	ID    int16
	Label string
}
