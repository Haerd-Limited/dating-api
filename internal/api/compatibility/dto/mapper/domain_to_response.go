package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
)

func MapDomainToGetOverviewResponse(d domain.Overview) dto.GetOverviewResponse {
	var packs []dto.Pack
	for _, pack := range d.QuestionPacks {
		packs = append(packs, dto.Pack{
			CategoryKey:                pack.CategoryKey,
			CategoryName:               pack.CategoryName,
			NumberOfCompletedQuestions: pack.NumberOfCompletedQuestions,
			TotalQuestions:             pack.TotalQuestions,
			ProgressPercent:            pack.ProgressPercent,
		})
	}

	return dto.GetOverviewResponse{
		QuestionPacks: packs,
	}
}

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
			IsAnswered:   qa.UserAnswer != nil,
		}
		q.Answers = make([]dto.AnswerResponse, 0, len(qa.Answers))

		for _, a := range qa.Answers {
			q.Answers = append(q.Answers, dto.AnswerResponse{
				ID:    a.ID,
				Label: a.Label,
				Sort:  a.Sort,
			})
		}

		// Map user answer if present
		if qa.UserAnswer != nil {
			q.UserAnswer = &dto.UserAnswerResponse{
				QuestionID:          qa.UserAnswer.QuestionID,
				AnswerID:            qa.UserAnswer.AnswerID,
				AcceptableAnswerIds: qa.UserAnswer.AcceptableAnswerIds,
				Importance:          qa.UserAnswer.Importance,
				IsPrivate:           qa.UserAnswer.IsPrivate,
				UpdatedAt:           qa.UserAnswer.UpdatedAt.Format(time.RFC3339),
			}
		}

		out.Questions = append(out.Questions, q)
	}

	// Map progress summary if present
	if d.ProgressSummary != nil {
		out.ProgressSummary = &dto.ProgressSummaryResponse{
			CategoryKey:                d.ProgressSummary.CategoryKey,
			CategoryName:               d.ProgressSummary.CategoryName,
			NumberOfCompletedQuestions: d.ProgressSummary.NumberOfCompletedQuestions,
			TotalQuestions:             d.ProgressSummary.TotalQuestions,
			ProgressPercent:            d.ProgressSummary.ProgressPercent,
			NextQuestionID:             d.ProgressSummary.NextQuestionID,
		}
	}

	return out
}
