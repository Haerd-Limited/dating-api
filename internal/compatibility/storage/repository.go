package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type CompatibilityRepository interface {
	ListQuestions(ctx context.Context, categoryKey *string, limit, offset int) (entity.QuestionSlice, error)
	GetQuestionAnswers(ctx context.Context, questionID int64) (entity.QuestionAnswerSlice, error)
	CountQuestions(ctx context.Context, categoryKey *string) (int, error)
	GetAnswerIDsForQuestion(ctx context.Context, questionID int64) ([]int64, error)
	UpsertUserAnswer(ctx context.Context, answer entity.UserAnswer) error
	GetQuestionCategories(ctx context.Context) (entity.QuestionCategorySlice, error)
	GetUserAnswers(ctx context.Context, userID string) (entity.UserAnswerSlice, error)
	CountAnsweredByCategory(ctx context.Context, userID, categoryKey string) (int, error)

	HasMandatoryMismatch(ctx context.Context, aID, bID string) (bool, error)
	PerspectiveSums(ctx context.Context, aID, bID string) (earned int, total int, overlap int, err error)
	TopBadges(ctx context.Context, aID, bID string, limit int) ([]domain.CompatibilityBadge, error)
	MandatoryMismatchBadges(ctx context.Context, viewerID, targetID string, limit int) ([]domain.CompatibilityBadge, error)

	// Sequential question methods
	GetNextUnansweredQuestion(ctx context.Context, userID, categoryKey string) (*entity.Question, error)
	GetUserAnsweredQuestionIDs(ctx context.Context, userID, categoryKey string) ([]int64, error)
	GetQuestionsInOrder(ctx context.Context, categoryKey string) (entity.QuestionSlice, error)
	GetUserAnswerForQuestion(ctx context.Context, userID string, questionID int64) (*entity.UserAnswer, error)
	GetQuestionSortOrderAndCategory(ctx context.Context, questionID int64) (sortOrder int, categoryKey string, err error)
	GetQuestionByIDAndCategory(ctx context.Context, questionID int64, categoryKey string) (*entity.Question, error)
}

type repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewCompatibilityRepository(
	db *sqlx.DB,
	logger *zap.Logger,
) CompatibilityRepository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

func (r *repository) GetUserAnswers(ctx context.Context, userID string) (entity.UserAnswerSlice, error) {
	result, err := entity.UserAnswers(
		entity.UserAnswerWhere.UserID.EQ(userID),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) GetQuestionCategories(ctx context.Context) (entity.QuestionCategorySlice, error) {
	result, err := entity.QuestionCategories().All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// HasMandatoryMismatch returns true if A has any 'mandatory' answers that B does not satisfy.
func (r *repository) HasMandatoryMismatch(ctx context.Context, aID, bID string) (bool, error) {
	const q = `
		SELECT 1
		FROM user_answers uaA
		JOIN user_answers uaB
		  ON uaB.user_id = $2
		 AND uaB.question_id = uaA.question_id
		WHERE uaA.user_id = $1
		  AND uaA.importance = 'mandatory'
		  AND NOT (uaB.answer_id = ANY(uaA.acceptable_answer_ids))
		LIMIT 1;
	`

	var one int

	err := queries.Raw(q, aID, bID).QueryRowContext(ctx, r.db).Scan(&one)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("HasMandatoryMismatch: %w", err)
	}

	return true, nil
}

// PerspectiveSums computes A's earned, total, and overlap versus B.
func (r *repository) PerspectiveSums(ctx context.Context, aID, bID string) (earned int, total int, overlap int, err error) {
	const q = `
		WITH overlap AS (
			SELECT uaA.importance,
			       uaA.acceptable_answer_ids AS acc,
			       uaB.answer_id             AS b_ans
			FROM user_answers uaA
			JOIN user_answers uaB
			  ON uaB.user_id = $2
			 AND uaB.question_id = uaA.question_id
			WHERE uaA.user_id = $1
		),
		w AS (SELECT key, weight FROM importance_weights)
		SELECT
			COALESCE(SUM(CASE WHEN b_ans = ANY(acc) THEN w.weight ELSE 0 END), 0)::int AS earned,
			COALESCE(SUM(w.weight), 0)::int                                             AS total,
			COUNT(*)::int                                                               AS overlap
		FROM overlap o
		JOIN w ON w.key = o.importance;
	`

	err = queries.Raw(q, aID, bID).QueryRowContext(ctx, r.db).Scan(&earned, &total, &overlap)
	if err != nil {
		err = fmt.Errorf("PerspectiveSums: %w", err)
	}

	return
}

// TopBadges returns A-perspective satisfied, highest-weight overlaps vs B.
func (r *repository) TopBadges(ctx context.Context, aID, bID string, limit int) ([]domain.CompatibilityBadge, error) {
	if limit <= 0 {
		limit = 3
	}

	const q = `
		SELECT q.id, q.text, qa.label, iw.weight
		FROM user_answers uaA
		JOIN user_answers uaB
		  ON uaB.user_id = $2
		 AND uaB.question_id = uaA.question_id
		JOIN questions q
		  ON q.id = uaA.question_id
		JOIN question_answers qa
		  ON qa.id = uaB.answer_id
		JOIN importance_weights iw
		  ON iw.key = uaA.importance
		WHERE uaA.user_id = $1
		  AND uaB.answer_id = ANY(uaA.acceptable_answer_ids)
		ORDER BY iw.weight DESC, q.id
		LIMIT $3;
	`

	type row struct {
		ID     int64
		Text   string
		Label  string
		Weight int
	}

	rows, err := queries.Raw(q, aID, bID, limit).QueryContext(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("TopBadges query: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			r.logger.Error("TopBadges close", zap.Error(err))
		}
	}(rows)

	var out []domain.CompatibilityBadge

	for rows.Next() {
		var rrow row
		if err = rows.Scan(&rrow.ID, &rrow.Text, &rrow.Label, &rrow.Weight); err != nil {
			return nil, fmt.Errorf("TopBadges scan: %w", err)
		}

		out = append(out, domain.CompatibilityBadge{
			QuestionID:    rrow.ID,
			QuestionText:  rrow.Text,
			PartnerAnswer: rrow.Label,
			Weight:        rrow.Weight,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("TopBadges rows: %w", err)
	}

	return out, nil
}

// MandatoryMismatchBadges returns badges for mandatory requirements that were not met (viewer's or target's), for UX to explain incompatibility.
func (r *repository) MandatoryMismatchBadges(ctx context.Context, viewerID, targetID string, limit int) ([]domain.CompatibilityBadge, error) {
	if limit <= 0 {
		limit = 3
	}

	const q = `
		SELECT id, text, label, requirement_by FROM (
			SELECT q.id, q.text, qa.label, 'viewer'::text AS requirement_by
			FROM user_answers uaV
			JOIN user_answers uaT ON uaT.user_id = $2 AND uaT.question_id = uaV.question_id
			JOIN questions q ON q.id = uaV.question_id
			JOIN question_answers qa ON qa.id = uaT.answer_id
			WHERE uaV.user_id = $1
			  AND uaV.importance = 'mandatory'
			  AND NOT (uaT.answer_id = ANY(uaV.acceptable_answer_ids))
			UNION ALL
			SELECT q.id, q.text, qa.label, 'target'::text AS requirement_by
			FROM user_answers uaT
			JOIN user_answers uaV ON uaV.user_id = $1 AND uaV.question_id = uaT.question_id
			JOIN questions q ON q.id = uaT.question_id
			JOIN question_answers qa ON qa.id = uaV.answer_id
			WHERE uaT.user_id = $2
			  AND uaT.importance = 'mandatory'
			  AND NOT (uaV.answer_id = ANY(uaT.acceptable_answer_ids))
		) AS u
		ORDER BY requirement_by, id
		LIMIT $3;
	`

	type row struct {
		ID            int64
		Text          string
		Label         string
		RequirementBy string
	}

	rows, err := queries.Raw(q, viewerID, targetID, limit).QueryContext(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("MandatoryMismatchBadges query: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	const mandatoryWeight = 30 // from importance_weights.mandatory

	var out []domain.CompatibilityBadge

	for rows.Next() {
		var rrow row
		if err = rows.Scan(&rrow.ID, &rrow.Text, &rrow.Label, &rrow.RequirementBy); err != nil {
			return nil, fmt.Errorf("MandatoryMismatchBadges scan: %w", err)
		}

		out = append(out, domain.CompatibilityBadge{
			QuestionID:    rrow.ID,
			QuestionText:  rrow.Text,
			PartnerAnswer: rrow.Label,
			Weight:        mandatoryWeight,
			IsMismatch:    true,
			RequirementBy: rrow.RequirementBy,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("MandatoryMismatchBadges rows: %w", err)
	}

	return out, nil
}

func (r *repository) GetAnswerIDsForQuestion(ctx context.Context, questionID int64) ([]int64, error) {
	ans, err := entity.QuestionAnswers(
		entity.QuestionAnswerWhere.QuestionID.EQ(questionID),
		qm.Select("id"),
		qm.OrderBy("sort ASC, id ASC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	out := make([]int64, 0, len(ans))
	for _, a := range ans {
		out = append(out, a.ID)
	}

	return out, nil
}

func (r *repository) UpsertUserAnswer(ctx context.Context, answer entity.UserAnswer) error {
	// Upsert:
	// - conflict on composite PK (user_id, question_id)
	// - update these columns on conflict
	// - insert uses inferred columns (or whitelist if you prefer)
	return answer.Upsert(
		ctx,
		r.db,
		true, // updateOnConflict
		[]string{entity.UserAnswerColumns.UserID, entity.UserAnswerColumns.QuestionID}, // conflict columns
		boil.Whitelist(
			entity.UserAnswerColumns.AnswerID,
			entity.UserAnswerColumns.AcceptableAnswerIds,
			entity.UserAnswerColumns.Importance,
			entity.UserAnswerColumns.IsPrivate,
			entity.UserAnswerColumns.UpdatedAt,
		), // columns to update on conflict
		boil.Infer(), // columns to insert (infer from struct)
	)
}

func (r *repository) GetQuestionAnswers(ctx context.Context, questionID int64) (entity.QuestionAnswerSlice, error) {
	answers, err := entity.QuestionAnswers(
		entity.QuestionAnswerWhere.QuestionID.EQ(questionID),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (r *repository) ListQuestions(ctx context.Context, categoryKey *string, limit, offset int) (entity.QuestionSlice, error) {
	// defensive defaults
	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	qmods := []qm.QueryMod{
		entity.QuestionWhere.IsActive.EQ(true),
		qm.InnerJoin("question_categories qc ON qc.id = questions.category_id"),
		qm.Load(entity.QuestionRels.Category),
		qm.Select("questions.*"),
		qm.OrderBy("questions.sort_order ASC"),
		qm.Limit(limit),
		qm.Offset(offset),
	}

	if categoryKey != nil && *categoryKey != "" {
		qmods = append(qmods, qm.Where("qc.key = ?", *categoryKey))
	}

	questions, err := entity.Questions(qmods...).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *repository) CountQuestions(ctx context.Context, categoryKey *string) (int, error) {
	qmods := []qm.QueryMod{
		entity.QuestionWhere.IsActive.EQ(true),
		qm.InnerJoin("question_categories qc ON qc.id = questions.category_id"),
	}

	if categoryKey != nil && *categoryKey != "" {
		qmods = append(qmods, qm.Where("qc.key = ?", *categoryKey))
	}

	count, err := entity.Questions(qmods...).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("count questions: %w", err)
	}

	return int(count), nil
}

// CountAnsweredByCategory returns how many questions `userID` has answered in `categoryKey`.
func (r *repository) CountAnsweredByCategory(ctx context.Context, userID, categoryKey string) (int, error) {
	const q = `
		SELECT COUNT(*)::int
		FROM user_answers ua
		JOIN questions q   ON q.id = ua.question_id
		JOIN question_categories qc ON qc.id = q.category_id
		WHERE ua.user_id = $1::uuid
		  AND qc.key = $2;
	`

	var n int
	if err := queries.Raw(q, userID, categoryKey).QueryRowContext(ctx, r.db).Scan(&n); err != nil {
		return 0, fmt.Errorf("CountAnsweredByCategory: %w", err)
	}

	return n, nil
}

// GetNextUnansweredQuestion returns the next question a user should answer in a category.
// Returns the first question if none answered, or the next unanswered question in sequence.
func (r *repository) GetNextUnansweredQuestion(ctx context.Context, userID, categoryKey string) (*entity.Question, error) {
	// Get all questions in order for the category
	allQuestions, err := r.GetQuestionsInOrder(ctx, categoryKey)
	if err != nil {
		return nil, fmt.Errorf("GetNextUnansweredQuestion: %w", err)
	}

	// Get answered question IDs for this user in this category
	answeredIDs, err := r.GetUserAnsweredQuestionIDs(ctx, userID, categoryKey)
	if err != nil {
		return nil, fmt.Errorf("GetNextUnansweredQuestion: %w", err)
	}

	// Create a set of answered IDs for quick lookup
	answeredSet := make(map[int64]struct{}, len(answeredIDs))
	for _, id := range answeredIDs {
		answeredSet[id] = struct{}{}
	}

	// Find the first unanswered question
	for _, q := range allQuestions {
		if _, answered := answeredSet[q.ID]; !answered {
			return q, nil
		}
	}

	// All questions answered
	return nil, nil
}

// GetUserAnsweredQuestionIDs returns all question IDs that a user has answered in a category.
func (r *repository) GetUserAnsweredQuestionIDs(ctx context.Context, userID, categoryKey string) ([]int64, error) {
	const q = `
		SELECT q.id
		FROM user_answers ua
		JOIN questions q ON q.id = ua.question_id
		JOIN question_categories qc ON qc.id = q.category_id
		WHERE ua.user_id = $1::uuid
		  AND qc.key = $2
		ORDER BY q.sort_order ASC;
	`

	rows, err := queries.Raw(q, userID, categoryKey).QueryContext(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("GetUserAnsweredQuestionIDs: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			r.logger.Error("GetUserAnsweredQuestionIDs close", zap.Error(err))
		}
	}(rows)

	var questionIDs []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("GetUserAnsweredQuestionIDs scan: %w", err)
		}

		questionIDs = append(questionIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetUserAnsweredQuestionIDs rows: %w", err)
	}

	return questionIDs, nil
}

// GetQuestionsInOrder returns all questions in a category ordered by sort_order.
func (r *repository) GetQuestionsInOrder(ctx context.Context, categoryKey string) (entity.QuestionSlice, error) {
	qmods := []qm.QueryMod{
		entity.QuestionWhere.IsActive.EQ(true),
		qm.InnerJoin("question_categories qc ON qc.id = questions.category_id"),
		qm.Load(entity.QuestionRels.Category),
		qm.Select("questions.*"),
		qm.Where("qc.key = ?", categoryKey),
		qm.OrderBy("questions.sort_order ASC"),
	}

	questions, err := entity.Questions(qmods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("GetQuestionsInOrder: %w", err)
	}

	return questions, nil
}

// GetUserAnswerForQuestion returns a user's answer for a specific question, if it exists.
func (r *repository) GetUserAnswerForQuestion(ctx context.Context, userID string, questionID int64) (*entity.UserAnswer, error) {
	answer, err := entity.UserAnswers(
		entity.UserAnswerWhere.UserID.EQ(userID),
		entity.UserAnswerWhere.QuestionID.EQ(questionID),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No answer exists
		}

		return nil, fmt.Errorf("GetUserAnswerForQuestion: %w", err)
	}

	return answer, nil
}

// GetQuestionSortOrderAndCategory returns the sort_order and category_key for a question.
func (r *repository) GetQuestionSortOrderAndCategory(ctx context.Context, questionID int64) (sortOrder int, categoryKey string, err error) {
	const q = `
		SELECT q.sort_order, qc.key
		FROM questions q
		JOIN question_categories qc ON qc.id = q.category_id
		WHERE q.id = $1;
	`

	err = queries.Raw(q, questionID).QueryRowContext(ctx, r.db).Scan(&sortOrder, &categoryKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("question %d not found", questionID)
		}

		return 0, "", fmt.Errorf("GetQuestionSortOrderAndCategory: %w", err)
	}

	return sortOrder, categoryKey, nil
}

// GetQuestionByIDAndCategory returns a question by ID if it exists, is active, and belongs to the given category.
// Returns (nil, nil) when the question is not found or does not belong to the category.
func (r *repository) GetQuestionByIDAndCategory(ctx context.Context, questionID int64, categoryKey string) (*entity.Question, error) {
	qmods := []qm.QueryMod{
		entity.QuestionWhere.ID.EQ(questionID),
		entity.QuestionWhere.IsActive.EQ(true),
		qm.InnerJoin("question_categories qc ON qc.id = questions.category_id"),
		qm.Load(entity.QuestionRels.Category),
		qm.Select("questions.*"),
		qm.Where("qc.key = ?", categoryKey),
	}

	question, err := entity.Questions(qmods...).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("GetQuestionByIDAndCategory: %w", err)
	}

	return question, nil
}
