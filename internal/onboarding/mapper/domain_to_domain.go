package mapper

import (
	compatibilitydomain "github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func MapCompatibilityOverviewToQuestionPacksContent(overview compatibilitydomain.Overview) domain.QuestionPacksContent {
	var questionPacks []domain.QuestionPack

	for _, pack := range overview.QuestionPacks {
		questionPacks = append(questionPacks, domain.QuestionPack{
			CategoryKey:                pack.CategoryKey,
			CategoryName:               pack.CategoryName,
			NumberOfCompletedQuestions: pack.NumberOfCompletedQuestions,
			TotalQuestions:             pack.TotalQuestions,
			ProgressPercent:            pack.ProgressPercent,
		})
	}

	return domain.QuestionPacksContent{QuestionPacks: questionPacks}
}
