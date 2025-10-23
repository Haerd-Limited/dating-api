package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/matching/dto"
	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
)

func MapDomainToQuestionAndAnswerResponse(d domain.QuestionsAndAnswers) dto.GetQuestionsAndAnswersResponse {
	out := dto.GetQuestionsAndAnswersResponse{
		Total:  d.Total,
		Limit:  d.Limit,
		Offset: d.Offset,
	}

	out.Questions = make([]dto.QuestionResponse, 0, len(d.Items))

	for _, qa := range d.Items {
		q := dto.QuestionResponse{
			ID:           qa.Question.ID,
			CategoryKey:  qa.Question.CategoryKey,
			CategoryName: qa.Question.CategoryName,
			Text:         qa.Question.Text,
		}
		q.Answers = make([]dto.AnswerResponse, 0, len(qa.Answers))

		for _, a := range qa.Answers {
			q.Answers = append(q.Answers, dto.AnswerResponse{
				ID:    a.ID,
				Label: a.Label,
				Sort:  a.Sort,
			})
		}

		out.Questions = append(out.Questions, q)
	}

	return out
}
