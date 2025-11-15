package domain

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}

type FamilyPlan struct {
	ID    int16
	Label string
}

type FamilyStatus struct {
	ID    int16
	Label string
}

type Language struct {
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

type Ethnicity struct {
	ID    int16
	Label string
}

type EducationLevel struct {
	ID    int16
	Label string
}

type ReportCategory struct {
	ID        int16
	Key       string
	Label     string
	SortOrder int16
}
