package domain

// Admin-side inputs and result rows for managing question packs
// (categories, questions, answer options) from the admin dashboard.

type CategoryInput struct {
	Key  string // only used on create; immutable afterwards
	Name string
}

type CategoryListRow struct {
	ID                  int64
	Key                 string
	Name                string
	SortOrder           int
	QuestionCount       int
	ActiveQuestionCount int
}

type QuestionInput struct {
	CategoryID int64
	Text       string
	IsActive   bool
}

type QuestionAdminRow struct {
	ID              int64
	CategoryID      int64
	Text            string
	IsActive        bool
	SortOrder       int
	AnswerCount     int
	UserAnswerCount int
}

type AnswerInput struct {
	QuestionID int64
	Label      string
}

type AnswerAdminRow struct {
	ID              int64
	QuestionID      int64
	Label           string
	Sort            int
	UserAnswerCount int
}

// ReorderCommand carries an ordered list of IDs for a single parent;
// the service renumbers them contiguously (1..N) server-side.
type ReorderCommand struct {
	OrderedIDs []int64
}
