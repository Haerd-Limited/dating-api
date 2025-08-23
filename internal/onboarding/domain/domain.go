package domain

type Register struct {
	Email          string
	PhoneNumber    string
	FirstName      string
	LastName       *string
	OnboardingStep Steps
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
