package domain

import "time"

type Basics struct {
	UserID            string
	Birthdate         time.Time
	HeightCm          int16
	GenderID          int16
	DatingIntentionID int16
}

type BasicsResult struct {
	OnboardingSteps
}

type Register struct {
	Email       string
	PhoneNumber string
	FirstName   string
	LastName    *string
}
type RegisterResult struct {
	AccessToken  string
	RefreshToken string
	OnboardingSteps
	Content RegisterContent
}

type OnboardingSteps struct {
	PreviousStep Steps
	CurrentStep  Steps
	NextStep     Steps
	Progress     float64
	Steps        []Steps
	TotalSteps   int
}

type RegisterContent struct {
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

type UserProfile struct {
	DisplayName *string
	Birthdate   time.Time
	HeightCM    int16
	UserID      string

	// Location
	Latitude  *float64
	Longitude *float64
	City      *string
	Country   *string

	// Single-selects
	GenderID          int16
	DatingIntentionID int16
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

	CreatedAt time.Time
	UpdatedAt time.Time
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
