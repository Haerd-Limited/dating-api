package dto

type AnswerResponse struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Sort  int    `json:"sort"`
}

type QuestionResponse struct {
	ID           int64            `json:"id"`
	CategoryKey  string           `json:"category_key"`
	CategoryName string           `json:"category_name"`
	Text         string           `json:"text"`
	Answers      []AnswerResponse `json:"answers"`
}

type GetOverviewResponse struct {
	QuestionPacks []Pack `json:"question_packs"`
}
type Pack struct {
	CategoryKey                string `json:"category_key"`
	CategoryName               string `json:"category_name"`
	NumberOfCompletedQuestions int    `json:"number_of_completed_questions"`
	TotalQuestions             int    `json:"total_questions"`
}

type GetQuestionsAndAnswersResponse struct {
	Questions []QuestionResponse `json:"questions"`
	Total     int                `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
	// NextOffset *int               `json:"next_offset,omitempty"`
	// PrevOffset *int               `json:"prev_offset,omitempty"`
	// HasMore    bool               `json:"has_more"`
}
