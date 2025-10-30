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
	Ethnicities     []string
	SpokenLanguages []string
	VoicePrompts    []VoicePrompt
	Theme           UserTheme

	Work       *string
	JobTitle   *string
	University *string

	Verified   bool
	DistanceKm int // todo(high-priority): implement logic

	LikeCount *int64

	MatchSummary *MatchSummary

	CreatedAt time.Time
	UpdatedAt time.Time
}

type MatchBadge struct {
	QuestionID    int64
	QuestionText  string
	PartnerAnswer string
	Weight        int // derived from importance
}

type MatchSummary struct {
	MatchPercent int          // 0–100
	OverlapCount int          // # shared questions answered
	Badges       []MatchBadge // top 2–3 satisfied items
	HiddenReason string       // e.g., "Not enough overlap"
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
