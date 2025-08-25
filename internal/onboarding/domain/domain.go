package domain

import "time"

type Basics struct {
	UserID            string
	Birthdate         time.Time
	HeightCm          int16
	GenderID          int16
	DatingIntentionID int16
}

type Background struct {
	UserID           string
	EducationLevelID int16
	EthnicityID      int16
}

type BackgroundResult struct {
	OnboardingSteps
}

type Beliefs struct {
	UserID             string
	PoliticalBeliefsID int16
	ReligionID         int16
}

type BeliefsResult struct {
	OnboardingSteps
	Content BeliefsContent
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

type LifestyleResult struct {
	OnboardingSteps
	Content LifestyleContent
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

type LocationResult struct {
	OnboardingSteps
	Content LocationContent
}

type LocationContent struct {
	Habits []Habit
}

type Habit struct {
	ID    int16
	Label string
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
	ChildrenStatusID  int16
	FamilyPlanID      *int32
	EthnicityID       int16

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
