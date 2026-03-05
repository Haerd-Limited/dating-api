package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func MapQuestionEntitiesToDomain(qe entity.QuestionSlice) []domain.Question {
	out := make([]domain.Question, 0, len(qe))

	for _, q := range qe {
		var catKey, catName string
		if q.R != nil && q.R.Category != nil {
			catKey = q.R.Category.Key
			catName = q.R.Category.Name
		}

		// Get sort_order from database if available (entity may not have it until regenerated)
		sortOrder := 0
		// Try to get it from a raw query or use ID as fallback
		// For now, we'll use 0 and rely on the repository ordering

		out = append(out, domain.Question{
			ID:           q.ID,
			CategoryKey:  catKey,
			CategoryName: catName,
			Text:         q.Text,
			IsActive:     q.IsActive,
			SortOrder:    sortOrder,
			CreatedAt:    q.CreatedAt,
		})
	}

	return out
}

func MapAnswerEntitiesToDomain(ae entity.QuestionAnswerSlice) []domain.AnswerOption {
	out := make([]domain.AnswerOption, 0, len(ae))
	for _, a := range ae {
		out = append(out, domain.AnswerOption{
			ID:         a.ID,
			QuestionID: a.QuestionID,
			Label:      a.Label,
			Sort:       a.Sort,
		})
	}

	return out
}

func MapQuestionEntityToDomain(q *entity.Question) domain.Question {
	var catKey, catName string
	if q.R != nil && q.R.Category != nil {
		catKey = q.R.Category.Key
		catName = q.R.Category.Name
	}

	// SortOrder will be 0 until entities are regenerated with the new column
	// The repository handles ordering by sort_order
	sortOrder := 0

	return domain.Question{
		ID:           q.ID,
		CategoryKey:  catKey,
		CategoryName: catName,
		Text:         q.Text,
		IsActive:     q.IsActive,
		SortOrder:    sortOrder,
		CreatedAt:    q.CreatedAt,
	}
}

func MapUserAnswerEntityToDomain(ua *entity.UserAnswer) *domain.UserAnswer {
	if ua == nil {
		return nil
	}

	return &domain.UserAnswer{
		QuestionID:          ua.QuestionID,
		AnswerID:            ua.AnswerID,
		AcceptableAnswerIds: ua.AcceptableAnswerIds,
		Importance:          ua.Importance,
		IsPrivate:           ua.IsPrivate,
		UpdatedAt:           ua.UpdatedAt,
	}
}
