package profilecard

import "time"

type ProfileCard struct {
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

	Gender          string
	DatingIntention string
	Sexuality       string
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

	VerifiedStatus string
	DistanceKm     int

	LikeCount *int64

	CompatibilitySummary *CompatibilitySummary

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CompatibilityBadge struct {
	QuestionID    int64
	QuestionText  string
	PartnerAnswer string
	Weight        int    // derived from importance
	IsMismatch    bool   // true when this badge describes a mandatory requirement that was not met
	RequirementBy string // when IsMismatch: "viewer" = viewer's mandatory that other failed; "target" = target's mandatory that viewer failed
}

type CompatibilitySummary struct {
	CompatibilityPercent int                  // 0–100
	OverlapCount         int                  // # shared questions answered
	Badges               []CompatibilityBadge // top 2–3 satisfied items
	HiddenReason         string               // e.g., "Not enough overlap"
}

type UserTheme struct {
	BaseHex string
	Palette []string
}

type VoicePrompt struct {
	ID                    int64
	URL                   string
	PromptType            Prompt
	IsPrimary             bool
	Position              int16
	DurationMs            int
	CoverMediaURL         string
	CoverMediaType        *string
	CoverMediaAspectRatio *float64
}

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}
