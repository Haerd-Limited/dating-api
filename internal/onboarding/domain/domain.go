package domain

type OnboardingUpdate struct {
	UserID string

	// Profile
	UserProfile *UserProfile

	Preferences *Preferences

	// Multi-selects (replace if provided)
	LanguageIDs *[]int32
	InterestIDs *[]int32

	// Progress
	BumpOnboardingStep *bool
}

type UserProfile struct {
	DisplayName *string
	Birthdate   *string // keep string; service can parse to time.Time if needed
	HeightCM    *int16

	// Location
	Latitude  *float64
	Longitude *float64
	City      *string
	Country   *string

	// Single-selects
	GenderID          *int32
	DatingIntentionID *int32
	ReligionID        *int32
	EducationLevelID  *int32
	PoliticalBeliefID *int32
	DrinkingID        *int32
	SmokingID         *int32
	MarijuanaID       *int32
	DrugsID           *int32
	ChildrenStatusID  *int32
	FamilyPlanID      *int32
	EthnicityID       *int32

	// Extra text fields in user_profiles
	Work        *string
	JobTitle    *string
	University  *string
	ProfileMeta *map[string]any
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
