package dto

type AdminCategoryResponse struct {
	ID                  int64  `json:"id"`
	Key                 string `json:"key"`
	Name                string `json:"name"`
	SortOrder           int    `json:"sort_order"`
	QuestionCount       int    `json:"question_count"`
	ActiveQuestionCount int    `json:"active_question_count"`
}

type AdminListCategoriesResponse struct {
	Categories []AdminCategoryResponse `json:"categories"`
}

type AdminQuestionResponse struct {
	ID              int64  `json:"id"`
	CategoryID      int64  `json:"category_id"`
	Text            string `json:"text"`
	IsActive        bool   `json:"is_active"`
	SortOrder       int    `json:"sort_order"`
	AnswerCount     int    `json:"answer_count"`
	UserAnswerCount int    `json:"user_answer_count"`
}

type AdminListQuestionsResponse struct {
	Questions []AdminQuestionResponse `json:"questions"`
}

type AdminAnswerResponse struct {
	ID              int64  `json:"id"`
	QuestionID      int64  `json:"question_id"`
	Label           string `json:"label"`
	Sort            int    `json:"sort"`
	UserAnswerCount int    `json:"user_answer_count"`
}

type AdminListAnswersResponse struct {
	Answers []AdminAnswerResponse `json:"answers"`
}
