package profilecard

import "time"

type ProfileCard struct {
	DisplayName   string
	Birthdate     time.Time
	Age           int
	HeightCM      int16
	UserID        string
	CoverPhotoUrl *string
	Emoji         string

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

	Work       *string
	JobTitle   *string
	University *string

	Verified   bool // todo: implement logic
	DistanceKm int  // todo: implement logic

	LikeCount *int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserTheme struct {
	BaseHex string
	Palette []string
}

type VoicePrompt struct {
	ID            int64
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
