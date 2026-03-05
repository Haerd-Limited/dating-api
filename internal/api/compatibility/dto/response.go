package dto

type AnswerResponse struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Sort  int    `json:"sort"`
}

type UserAnswerResponse struct {
	QuestionID          int64   `json:"question_id"`
	AnswerID            int64   `json:"answer_id"`
	AcceptableAnswerIds []int64 `json:"acceptable_answer_ids"`
	Importance          string  `json:"importance"`
	IsPrivate           bool    `json:"is_private"`
	UpdatedAt           string  `json:"updated_at"`
}

type QuestionResponse struct {
	ID           int64               `json:"id"`
	CategoryKey  string              `json:"category_key"`
	CategoryName string              `json:"category_name"`
	Text         string              `json:"text"`
	Answers      []AnswerResponse    `json:"answers"`
	UserAnswer   *UserAnswerResponse `json:"user_answer,omitempty"`
	IsAnswered   bool                `json:"is_answered"`
}

type GetOverviewResponse struct {
	QuestionPacks []Pack `json:"question_packs"`
}
type Pack struct {
	CategoryKey                string  `json:"category_key"`
	CategoryName               string  `json:"category_name"`
	NumberOfCompletedQuestions int     `json:"number_of_completed_questions"`
	TotalQuestions             int     `json:"total_questions"`
	ProgressPercent            float64 `json:"progress_percent"`
}

type ProgressSummaryResponse struct {
	CategoryKey                string  `json:"category_key"`
	CategoryName               string  `json:"category_name"`
	NumberOfCompletedQuestions int     `json:"number_of_completed_questions"`
	TotalQuestions             int     `json:"total_questions"`
	ProgressPercent            float64 `json:"progress_percent"`
	NextQuestionID             *int64  `json:"next_question_id,omitempty"` // ID of next unanswered question, nil if all answered
}

type GetQuestionsAndAnswersResponse struct {
	Questions       []QuestionResponse       `json:"questions"`
	Total           int                      `json:"total"`
	Limit           int                      `json:"limit"`
	Offset          int                      `json:"offset"`
	ProgressSummary *ProgressSummaryResponse `json:"progress_summary,omitempty"` // Only present when view=all
	// NextOffset *int               `json:"next_offset,omitempty"`
	// PrevOffset *int               `json:"prev_offset,omitempty"`
	// HasMore    bool               `json:"has_more"`
}
