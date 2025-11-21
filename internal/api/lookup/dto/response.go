package dto

type GetPromptsResponse struct {
	Prompts []Prompt `json:"prompts"`
}

type GetFamilyStatusResponse struct {
	FamilyStatus []FamilyStatus `json:"family_status"`
}

type FamilyStatus struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetFamilyPlansResponse struct {
	FamilyPlans []FamilyPlan `json:"family_plans"`
}

type FamilyPlan struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

type GetLanguagesResponse struct {
	Languages []Language `json:"languages"`
}

type Language struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetReligionsResponse struct {
	Religions []Religion `json:"religions"`
}

type Religion struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetPoliticalBeliefsResponse struct {
	PoliticalBeliefs []PoliticalBelief `json:"political_beliefs"`
}

type PoliticalBelief struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetHabitsResponse struct {
	Habits []Habit `json:"habits"`
}

type Habit struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetGendersResponse struct {
	Genders []Gender `json:"genders"`
}

type Gender struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetSexualitiesResponse struct {
	Sexualities []Sexuality `json:"sexualities"`
}

type Sexuality struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetDatingIntentionsResponse struct {
	DatingIntentions []DatingIntention `json:"dating_intentions"`
}

type DatingIntention struct {
	ID          int16  `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type GetEthnicitiesResponse struct {
	Ethnicities []Ethnicity `json:"ethnicities"`
}

type Ethnicity struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetEducationLevelsResponse struct {
	EducationLevels []EducationLevel `json:"education_levels"`
}

type EducationLevel struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetReportCategoriesResponse struct {
	ReportCategories []ReportCategory `json:"report_categories"`
}

type ReportCategory struct {
	ID        int16  `json:"id"`
	Key       string `json:"key"`
	Label     string `json:"label"`
	SortOrder int16  `json:"sort_order"`
}
