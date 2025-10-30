package dto

type ProfileCard struct {
	DisplayName string `json:"display_name"`
	Birthdate   string `json:"birthdate"`
	Age         int    `json:"age"`
	HeightCM    int16  `json:"height_cm"`
	UserID      string `json:"user_id"`
	Emoji       string `json:"emoji"`

	// Location
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"home_town"`
	Country   string  `json:"country"`

	Gender          string        `json:"gender"`
	DatingIntention string        `json:"dating_intention"`
	Religion        string        `json:"religion"`
	EducationLevel  string        `json:"education_level"`
	PoliticalBelief string        `json:"political_belief"`
	Drinking        string        `json:"drinking"`
	Smoking         string        `json:"smoking"`
	Marijuana       string        `json:"marijuana"`
	Drugs           string        `json:"drugs"`
	ChildrenStatus  *string       `json:"children_status"`
	FamilyPlan      *string       `json:"family_plan"`
	Ethnicities     []string      `json:"ethnicities"`
	SpokenLanguages []string      `json:"spoken_languages"`
	VoicePrompts    []VoicePrompt `json:"voice_prompts"`
	Theme           UserTheme     `json:"theme"`
	CoverPhotoURL   string        `json:"cover_photo_url"`

	Verified  bool   `json:"verified"`
	LikeCount *int64 `json:"like_count"`

	Work         *string       `json:"work"`
	JobTitle     *string       `json:"job_title"`
	University   *string       `json:"university"`
	MatchSummary *MatchSummary `json:"match_summary"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type MatchBadge struct {
	QuestionID    int64  `json:"question_id"`
	QuestionText  string `json:"question_text"`
	PartnerAnswer string `json:"partner_answer"`
	Weight        int    `json:"weight"` // derived from importance
}

type MatchSummary struct {
	MatchPercent int          `json:"match_percent"`           // 0–100
	OverlapCount int          `json:"overlap_count"`           // # shared questions answered
	Badges       []MatchBadge `json:"badges"`                  // top 2–3 satisfied items
	HiddenReason string       `json:"hidden_reason,omitempty"` // e.g., "Not enough overlap"
}
type UserTheme struct {
	BaseHex string   `json:"base_hex"`
	Palette []string `json:"palette"`
}

type VoicePrompt struct {
	ID            int64  `json:"id"`
	URL           string `json:"url"`
	PromptType    Prompt `json:"prompt_type"`
	IsPrimary     bool   `json:"is_primary"`
	Position      int16  `json:"position"`
	DurationMs    int    `json:"duration_ms"`
	CoverPhotoUrl string `json:"cover_photo_url"`
}

type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
