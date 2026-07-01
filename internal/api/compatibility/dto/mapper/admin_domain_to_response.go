package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
)

func MapCategoryRowsToResponse(rows []domain.CategoryListRow) dto.AdminListCategoriesResponse {
	out := make([]dto.AdminCategoryResponse, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.AdminCategoryResponse{
			ID:                  c.ID,
			Key:                 c.Key,
			Name:                c.Name,
			SortOrder:           c.SortOrder,
			QuestionCount:       c.QuestionCount,
			ActiveQuestionCount: c.ActiveQuestionCount,
		})
	}

	return dto.AdminListCategoriesResponse{Categories: out}
}

func MapCategoryRowToResponse(c domain.CategoryListRow) dto.AdminCategoryResponse {
	return dto.AdminCategoryResponse{
		ID:                  c.ID,
		Key:                 c.Key,
		Name:                c.Name,
		SortOrder:           c.SortOrder,
		QuestionCount:       c.QuestionCount,
		ActiveQuestionCount: c.ActiveQuestionCount,
	}
}

func MapQuestionRowsToResponse(rows []domain.QuestionAdminRow) dto.AdminListQuestionsResponse {
	out := make([]dto.AdminQuestionResponse, 0, len(rows))
	for _, q := range rows {
		out = append(out, MapQuestionRowToResponse(q))
	}

	return dto.AdminListQuestionsResponse{Questions: out}
}

func MapQuestionRowToResponse(q domain.QuestionAdminRow) dto.AdminQuestionResponse {
	return dto.AdminQuestionResponse{
		ID:              q.ID,
		CategoryID:      q.CategoryID,
		Text:            q.Text,
		IsActive:        q.IsActive,
		SortOrder:       q.SortOrder,
		AnswerCount:     q.AnswerCount,
		UserAnswerCount: q.UserAnswerCount,
	}
}

func MapAnswerRowsToResponse(rows []domain.AnswerAdminRow) dto.AdminListAnswersResponse {
	out := make([]dto.AdminAnswerResponse, 0, len(rows))
	for _, a := range rows {
		out = append(out, MapAnswerRowToResponse(a))
	}

	return dto.AdminListAnswersResponse{Answers: out}
}

func MapAnswerRowToResponse(a domain.AnswerAdminRow) dto.AdminAnswerResponse {
	return dto.AdminAnswerResponse{
		ID:              a.ID,
		QuestionID:      a.QuestionID,
		Label:           a.Label,
		Sort:            a.Sort,
		UserAnswerCount: a.UserAnswerCount,
	}
}
