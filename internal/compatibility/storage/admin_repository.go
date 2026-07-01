package storage

//go:generate mockgen -source=admin_repository.go -destination=admin_repository_mock.go -package=storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

// Storage-level sentinel errors. The admin service classifies these into
// domain-specific errors (duplicate key/text/label, category-not-empty, etc.).
var (
	ErrNotFound            = errors.New("not found")
	ErrUniqueViolation     = errors.New("unique violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// AdminCompatibilityRepository provides admin CRUD over the question-pack
// content tables (categories, questions, answer options).
type AdminCompatibilityRepository interface {
	// categories
	ListCategoriesWithCounts(ctx context.Context) ([]domain.CategoryListRow, error)
	GetCategory(ctx context.Context, id int64) (*entity.QuestionCategory, error)
	CreateCategory(ctx context.Context, key, name string, sortOrder int) (*entity.QuestionCategory, error)
	UpdateCategoryName(ctx context.Context, id int64, name string) error
	DeleteCategory(ctx context.Context, id int64) error
	ReorderCategories(ctx context.Context, orderedIDs []int64) error
	NextCategorySortOrder(ctx context.Context) (int, error)
	CountQuestionsInCategory(ctx context.Context, categoryID int64) (int, error)

	// questions
	ListQuestionsByCategoryAdmin(ctx context.Context, categoryID int64) ([]domain.QuestionAdminRow, error)
	GetQuestionAdmin(ctx context.Context, id int64) (*entity.Question, error)
	CreateQuestion(ctx context.Context, categoryID int64, text string, isActive bool, sortOrder int) (*entity.Question, error)
	UpdateQuestion(ctx context.Context, id int64, text string, isActive bool) error
	DeleteQuestion(ctx context.Context, id int64) error
	ReorderQuestions(ctx context.Context, categoryID int64, orderedIDs []int64) error
	NextQuestionSortOrder(ctx context.Context, categoryID int64) (int, error)
	CountUserAnswersForQuestion(ctx context.Context, questionID int64) (int, error)

	// answers
	ListAnswersByQuestionAdmin(ctx context.Context, questionID int64) ([]domain.AnswerAdminRow, error)
	GetAnswerAdmin(ctx context.Context, id int64) (*entity.QuestionAnswer, error)
	CreateAnswer(ctx context.Context, questionID int64, label string, sort int) (*entity.QuestionAnswer, error)
	UpdateAnswerLabel(ctx context.Context, id int64, label string) error
	DeleteAnswer(ctx context.Context, id int64) error
	ReorderAnswers(ctx context.Context, questionID int64, orderedIDs []int64) error
	NextAnswerSort(ctx context.Context, questionID int64) (int, error)
	CountUserAnswersForAnswer(ctx context.Context, answerID int64) (int, error)
}

type adminRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewAdminCompatibilityRepository(
	db *sqlx.DB,
	logger *zap.Logger,
) AdminCompatibilityRepository {
	return &adminRepository{
		db:     db,
		logger: logger,
	}
}

// translatePqError maps Postgres constraint violations to storage sentinels.
func translatePqError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return ErrUniqueViolation
		case "23503": // foreign_key_violation
			return ErrForeignKeyViolation
		}
	}

	return err
}

// ---- categories ----

func (r *adminRepository) ListCategoriesWithCounts(ctx context.Context) ([]domain.CategoryListRow, error) {
	const q = `
		SELECT qc.id, qc.key, qc.name, qc.sort_order,
		       COUNT(que.id)::int AS question_count,
		       COUNT(que.id) FILTER (WHERE que.is_active)::int AS active_question_count
		FROM question_categories qc
		LEFT JOIN questions que ON que.category_id = qc.id
		GROUP BY qc.id
		ORDER BY qc.sort_order ASC, qc.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListCategoriesWithCounts: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []domain.CategoryListRow

	for rows.Next() {
		var row domain.CategoryListRow
		if scanErr := rows.Scan(&row.ID, &row.Key, &row.Name, &row.SortOrder, &row.QuestionCount, &row.ActiveQuestionCount); scanErr != nil {
			return nil, fmt.Errorf("ListCategoriesWithCounts scan: %w", scanErr)
		}

		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListCategoriesWithCounts rows: %w", err)
	}

	return out, nil
}

func (r *adminRepository) GetCategory(ctx context.Context, id int64) (*entity.QuestionCategory, error) {
	c, err := entity.FindQuestionCategory(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("GetCategory: %w", err)
	}

	return c, nil
}

func (r *adminRepository) CreateCategory(ctx context.Context, key, name string, sortOrder int) (*entity.QuestionCategory, error) {
	c := &entity.QuestionCategory{Key: key, Name: name, SortOrder: sortOrder}
	if err := c.Insert(ctx, r.db, boil.Infer()); err != nil {
		return nil, translatePqError(err)
	}

	return c, nil
}

func (r *adminRepository) UpdateCategoryName(ctx context.Context, id int64, name string) error {
	c, err := r.GetCategory(ctx, id)
	if err != nil {
		return err
	}

	c.Name = name
	if _, err := c.Update(ctx, r.db, boil.Whitelist(entity.QuestionCategoryColumns.Name)); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) DeleteCategory(ctx context.Context, id int64) error {
	c, err := r.GetCategory(ctx, id)
	if err != nil {
		return err
	}

	if _, err := c.Delete(ctx, r.db); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) ReorderCategories(ctx context.Context, orderedIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ReorderCategories begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	const q = `UPDATE question_categories SET sort_order = $1 WHERE id = $2;`

	for i, id := range orderedIDs {
		res, execErr := tx.ExecContext(ctx, q, i+1, id)
		if execErr != nil {
			return fmt.Errorf("ReorderCategories update: %w", execErr)
		}

		affected, _ := res.RowsAffected()
		if affected != 1 {
			return ErrNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ReorderCategories commit: %w", err)
	}

	return nil
}

func (r *adminRepository) NextCategorySortOrder(ctx context.Context) (int, error) {
	var next int

	const q = `SELECT COALESCE(MAX(sort_order), 0)::int FROM question_categories;`

	if err := queries.Raw(q).QueryRowContext(ctx, r.db).Scan(&next); err != nil {
		return 0, fmt.Errorf("NextCategorySortOrder: %w", err)
	}

	return next, nil
}

func (r *adminRepository) CountQuestionsInCategory(ctx context.Context, categoryID int64) (int, error) {
	count, err := entity.Questions(entity.QuestionWhere.CategoryID.EQ(categoryID)).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("CountQuestionsInCategory: %w", err)
	}

	return int(count), nil
}

// ---- questions ----

func (r *adminRepository) ListQuestionsByCategoryAdmin(ctx context.Context, categoryID int64) ([]domain.QuestionAdminRow, error) {
	const q = `
		SELECT que.id, que.category_id, que.text, que.is_active, que.sort_order,
		       COUNT(DISTINCT qa.id)::int AS answer_count,
		       COUNT(DISTINCT ua.user_id)::int AS user_answer_count
		FROM questions que
		LEFT JOIN question_answers qa ON qa.question_id = que.id
		LEFT JOIN user_answers ua ON ua.question_id = que.id
		WHERE que.category_id = $1
		GROUP BY que.id
		ORDER BY que.sort_order ASC, que.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, categoryID)
	if err != nil {
		return nil, fmt.Errorf("ListQuestionsByCategoryAdmin: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []domain.QuestionAdminRow

	for rows.Next() {
		var row domain.QuestionAdminRow
		if scanErr := rows.Scan(&row.ID, &row.CategoryID, &row.Text, &row.IsActive, &row.SortOrder, &row.AnswerCount, &row.UserAnswerCount); scanErr != nil {
			return nil, fmt.Errorf("ListQuestionsByCategoryAdmin scan: %w", scanErr)
		}

		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListQuestionsByCategoryAdmin rows: %w", err)
	}

	return out, nil
}

func (r *adminRepository) GetQuestionAdmin(ctx context.Context, id int64) (*entity.Question, error) {
	q, err := entity.FindQuestion(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("GetQuestionAdmin: %w", err)
	}

	return q, nil
}

func (r *adminRepository) CreateQuestion(ctx context.Context, categoryID int64, text string, isActive bool, sortOrder int) (*entity.Question, error) {
	q := &entity.Question{
		CategoryID: categoryID,
		Text:       text,
		Type:       "structured",
		IsActive:   isActive,
		SortOrder:  sortOrder,
	}
	if err := q.Insert(ctx, r.db, boil.Infer()); err != nil {
		return nil, translatePqError(err)
	}

	return q, nil
}

func (r *adminRepository) UpdateQuestion(ctx context.Context, id int64, text string, isActive bool) error {
	q, err := r.GetQuestionAdmin(ctx, id)
	if err != nil {
		return err
	}

	q.Text = text
	q.IsActive = isActive

	if _, err := q.Update(ctx, r.db, boil.Whitelist(entity.QuestionColumns.Text, entity.QuestionColumns.IsActive)); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) DeleteQuestion(ctx context.Context, id int64) error {
	q, err := r.GetQuestionAdmin(ctx, id)
	if err != nil {
		return err
	}

	if _, err := q.Delete(ctx, r.db); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) ReorderQuestions(ctx context.Context, categoryID int64, orderedIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ReorderQuestions begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	const q = `UPDATE questions SET sort_order = $1 WHERE id = $2 AND category_id = $3;`

	for i, id := range orderedIDs {
		res, execErr := tx.ExecContext(ctx, q, i+1, id, categoryID)
		if execErr != nil {
			return fmt.Errorf("ReorderQuestions update: %w", execErr)
		}

		affected, _ := res.RowsAffected()
		if affected != 1 {
			return ErrNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ReorderQuestions commit: %w", err)
	}

	return nil
}

func (r *adminRepository) NextQuestionSortOrder(ctx context.Context, categoryID int64) (int, error) {
	var next int

	const q = `SELECT COALESCE(MAX(sort_order), 0)::int FROM questions WHERE category_id = $1;`

	if err := queries.Raw(q, categoryID).QueryRowContext(ctx, r.db).Scan(&next); err != nil {
		return 0, fmt.Errorf("NextQuestionSortOrder: %w", err)
	}

	return next, nil
}

func (r *adminRepository) CountUserAnswersForQuestion(ctx context.Context, questionID int64) (int, error) {
	var count int

	const q = `SELECT COUNT(*)::int FROM user_answers WHERE question_id = $1;`

	if err := queries.Raw(q, questionID).QueryRowContext(ctx, r.db).Scan(&count); err != nil {
		return 0, fmt.Errorf("CountUserAnswersForQuestion: %w", err)
	}

	return count, nil
}

// ---- answers ----

func (r *adminRepository) ListAnswersByQuestionAdmin(ctx context.Context, questionID int64) ([]domain.AnswerAdminRow, error) {
	const q = `
		SELECT qa.id, qa.question_id, qa.label, qa.sort,
		       (SELECT COUNT(*)::int FROM user_answers ua WHERE ua.answer_id = qa.id) AS user_answer_count
		FROM question_answers qa
		WHERE qa.question_id = $1
		ORDER BY qa.sort ASC, qa.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, questionID)
	if err != nil {
		return nil, fmt.Errorf("ListAnswersByQuestionAdmin: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []domain.AnswerAdminRow

	for rows.Next() {
		var row domain.AnswerAdminRow
		if scanErr := rows.Scan(&row.ID, &row.QuestionID, &row.Label, &row.Sort, &row.UserAnswerCount); scanErr != nil {
			return nil, fmt.Errorf("ListAnswersByQuestionAdmin scan: %w", scanErr)
		}

		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAnswersByQuestionAdmin rows: %w", err)
	}

	return out, nil
}

func (r *adminRepository) GetAnswerAdmin(ctx context.Context, id int64) (*entity.QuestionAnswer, error) {
	a, err := entity.FindQuestionAnswer(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("GetAnswerAdmin: %w", err)
	}

	return a, nil
}

func (r *adminRepository) CreateAnswer(ctx context.Context, questionID int64, label string, sort int) (*entity.QuestionAnswer, error) {
	a := &entity.QuestionAnswer{QuestionID: questionID, Label: label, Sort: sort}
	if err := a.Insert(ctx, r.db, boil.Infer()); err != nil {
		return nil, translatePqError(err)
	}

	return a, nil
}

func (r *adminRepository) UpdateAnswerLabel(ctx context.Context, id int64, label string) error {
	a, err := r.GetAnswerAdmin(ctx, id)
	if err != nil {
		return err
	}

	a.Label = label
	if _, err := a.Update(ctx, r.db, boil.Whitelist(entity.QuestionAnswerColumns.Label)); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) DeleteAnswer(ctx context.Context, id int64) error {
	a, err := r.GetAnswerAdmin(ctx, id)
	if err != nil {
		return err
	}

	if _, err := a.Delete(ctx, r.db); err != nil {
		return translatePqError(err)
	}

	return nil
}

func (r *adminRepository) ReorderAnswers(ctx context.Context, questionID int64, orderedIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ReorderAnswers begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	const q = `UPDATE question_answers SET sort = $1 WHERE id = $2 AND question_id = $3;`

	for i, id := range orderedIDs {
		res, execErr := tx.ExecContext(ctx, q, i+1, id, questionID)
		if execErr != nil {
			return fmt.Errorf("ReorderAnswers update: %w", execErr)
		}

		affected, _ := res.RowsAffected()
		if affected != 1 {
			return ErrNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ReorderAnswers commit: %w", err)
	}

	return nil
}

func (r *adminRepository) NextAnswerSort(ctx context.Context, questionID int64) (int, error) {
	var next int

	const q = `SELECT COALESCE(MAX(sort), 0)::int FROM question_answers WHERE question_id = $1;`

	if err := queries.Raw(q, questionID).QueryRowContext(ctx, r.db).Scan(&next); err != nil {
		return 0, fmt.Errorf("NextAnswerSort: %w", err)
	}

	return next, nil
}

func (r *adminRepository) CountUserAnswersForAnswer(ctx context.Context, answerID int64) (int, error) {
	var count int

	const q = `SELECT COUNT(*)::int FROM user_answers WHERE answer_id = $1;`

	if err := queries.Raw(q, answerID).QueryRowContext(ctx, r.db).Scan(&count); err != nil {
		return 0, fmt.Errorf("CountUserAnswersForAnswer: %w", err)
	}

	return count, nil
}
