package domain

import "time"

type Overview struct {
	QuestionPacks []Pack
}

type Pack struct {
	CategoryKey                string
	CategoryName               string
	NumberOfCompletedQuestions int
	TotalQuestions             int
	ProgressPercent            float64
}
type CompatibilityBadge struct {
	QuestionID    int64
	QuestionText  string
	PartnerAnswer string
	Weight        int // derived from importance
}

type CompatibilitySummary struct {
	CompatibilityPercent int                  // 0–100
	OverlapCount         int                  // # shared questions answered
	Badges               []CompatibilityBadge // top 2–3 satisfied items
	HiddenReason         string               // e.g., "Not enough overlap"
}

type SaveAnswerCommand struct {
	UserID              string
	QuestionID          int64
	AnswerID            int64
	AcceptableAnswerIDs []int64
	Importance          string // enum text
	IsPrivate           bool
}

type QuestionCategory struct {
	ID        int64
	Key       string
	Name      string
	CreatedAt time.Time
}

type Question struct {
	ID           int64
	CategoryKey  string
	CategoryName string
	Text         string
	IsActive     bool
	SortOrder    int
	CreatedAt    time.Time
}

type AnswerOption struct {
	ID         int64
	QuestionID int64
	Label      string
	Sort       int
}

type UserAnswer struct {
	QuestionID          int64
	AnswerID            int64
	AcceptableAnswerIds []int64
	Importance          string
	IsPrivate           bool
	UpdatedAt           time.Time
}

type QuestionAndAnswers struct {
	Question   Question
	Answers    []AnswerOption
	UserAnswer *UserAnswer // Optional: existing answer if question was previously answered
}

type QuestionsAndAnswers struct {
	Items  []QuestionAndAnswers
	Total  int // total rows matching the filter (for UI page counts)
	Limit  int // request limit
	Offset int // request offset
	// Progress summary (only populated when viewAll=true)
	ProgressSummary *ProgressSummary
	// NextOffset *int // nil if no more
	// PrevOffset *int // nil if already at 0
	// HasMore    bool // convenience
}

type ProgressSummary struct {
	CategoryKey                string
	CategoryName               string
	NumberOfCompletedQuestions int
	TotalQuestions             int
	ProgressPercent            float64
	NextQuestionID             *int64 // ID of next unanswered question, nil if all answered
}
