package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
)

func MapQuestionEntitiesToDomain(qe entity.QuestionSlice) []domain.Question {
	out := make([]domain.Question, 0, len(qe))

	for _, q := range qe {
		var catKey, catName string
		if q.R != nil && q.R.Category != nil {
			catKey = q.R.Category.Key
			catName = q.R.Category.Name
		}

		out = append(out, domain.Question{
			ID:           q.ID,
			CategoryKey:  catKey,
			CategoryName: catName,
			Text:         q.Text,
			IsActive:     q.IsActive,
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
