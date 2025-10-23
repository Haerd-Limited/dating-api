package storage

import (
	"context"
	"fmt"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"
)

type MatchingRepository interface {
	ListQuestions(ctx context.Context, categoryKey *string, limit, offset int) (entity.QuestionSlice, error)
	GetQuestionAnswers(ctx context.Context, questionID int64) (entity.QuestionAnswerSlice, error)
	CountQuestions(ctx context.Context, categoryKey *string) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewMatchingRepository(db *sqlx.DB) MatchingRepository {
	return &repository{
		db: db,
	}
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
		qm.OrderBy("questions.id ASC"),
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
