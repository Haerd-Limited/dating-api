package profilecard

import "time"

type ProfileCard struct {
	DisplayName string
	Birthdate   time.Time
	Age         int
	HeightCM    int16
	UserID      string

	// Location
	Latitude  float64
	Longitude float64
	City      string
	Country   string

	Gender          string
	DatingIntention string
	Religion        string
	EducationLevel  string
	PoliticalBelief string
	Drinking        string
	Smoking         string
	Marijuana       string
	Drugs           string
	ChildrenStatus  *string
	FamilyPlan      *string
	Ethnicity       string
	SpokenLanguages []string
	VoicePrompts    []VoicePrompt
	Theme           UserTheme
	CoverPhotoUrl   *string

	Work       *string
	JobTitle   *string
	University *string

	Verified   bool // todo: implement logic
	DistanceKm int  // todo: implement logic

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserTheme struct {
	BaseHex string
	Palette []string
}

type VoicePrompt struct {
	URL           string
	PromptType    Prompt
	IsPrimary     bool
	Position      int16
	DurationMs    int
	CoverPhotoUrl string
}

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}
