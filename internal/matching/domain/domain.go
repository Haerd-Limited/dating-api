package domain

import "time"

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
	CreatedAt    time.Time
}

type AnswerOption struct {
	ID         int64
	QuestionID int64
	Label      string
	Sort       int
}

type QuestionAndAnswers struct {
	Question Question
	Answers  []AnswerOption
}

type QuestionsAndAnswers struct {
	Items  []QuestionAndAnswers
	Total  int // total rows matching the filter (for UI page counts)
	Limit  int // request limit
	Offset int // request offset
	// NextOffset *int // nil if no more
	// PrevOffset *int // nil if already at 0
	// HasMore    bool // convenience
}
