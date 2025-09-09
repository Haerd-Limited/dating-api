package domain

import "time"

type Profile struct {
	DisplayName *string
	Birthdate   time.Time
	HeightCM    int16
	UserID      string

	// Location
	Latitude  float64
	Longitude float64
	City      string
	Country   string

	// Single-selects
	GenderID          int16
	DatingIntentionID int16
	ReligionID        int16
	EducationLevelID  int16
	PoliticalBeliefID int16
	DrinkingID        int16
	SmokingID         int16
	MarijuanaID       int16
	DrugsID           int16
	ChildrenStatusID  *int16
	FamilyPlanID      *int16
	EthnicityID       int16

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
	DisplayName *string
	Birthdate   *time.Time
	HeightCM    *int16
	UserID      string

	// Location
	Latitude  *float64
	Longitude *float64
	City      *string
	Country   *string

	// Single-selects
	GenderID          *int16
	DatingIntentionID *int16
	ReligionID        *int16
	EducationLevelID  *int16
	PoliticalBeliefID *int16
	DrinkingID        *int16
	SmokingID         *int16
	MarijuanaID       *int16
	DrugsID           *int16
	ChildrenStatusID  *int16
	FamilyPlanID      *int16
	EthnicityID       *int16
	SpokenLanguages   []int16
	VoicePrompts      []VoicePrompt
	Photos            []Photo

	Work       *string
	JobTitle   *string
	University *string

	CreatedAt *time.Time
	UpdatedAt time.Time
}

type EnrichedProfile struct {
	DisplayName *string
	Birthdate   time.Time
	Age         int
	HeightCM    int16
	UserID      string

	// Location
	Latitude  float64
	Longitude float64
	City      string
	Country   string

	Gender          Gender
	DatingIntention DatingIntention
	Religion        Religion
	EducationLevel  EducationLevel
	PoliticalBelief PoliticalBelief
	Drinking        Habit
	Smoking         Habit
	Marijuana       Habit
	Drugs           Habit
	ChildrenStatus  *Status
	FamilyPlan      *Status
	Ethnicity       Ethnicity
	SpokenLanguages []Language
	VoicePrompts    []VoicePrompt
	Photos          []Photo

	Work       *string
	JobTitle   *string
	University *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type VoicePrompt struct {
	URL            string
	PromptType     Prompt
	IsPrimary      bool
	Position       int16
	DurationMs     int
	PromptCoverURL string // TODO: HAVE frontend provide this for BE to store
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
