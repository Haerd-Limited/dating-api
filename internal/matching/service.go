package matching

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
	"github.com/Haerd-Limited/dating-api/internal/matching/mapper"
	"github.com/Haerd-Limited/dating-api/internal/matching/storage"
)

type Service interface {
	GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int) (domain.QuestionsAndAnswers, error)
}

type service struct {
	logger       *zap.Logger
	matchingRepo storage.MatchingRepository
}

func NewMatchingService(
	logger *zap.Logger,
	matchingRepository storage.MatchingRepository,
) Service {
	return &service{
		logger:       logger,
		matchingRepo: matchingRepository,
	}
}

func (s *service) GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int) (domain.QuestionsAndAnswers, error) {
	var catPtr *string
	if category != "" {
		catPtr = &category
	}

	questionEntities, err := s.matchingRepo.ListQuestions(ctx, catPtr, limit, offset)
	if err != nil {
		return domain.QuestionsAndAnswers{}, err
	}

	domainQuestion := mapper.MapQuestionEntitiesToDomain(questionEntities)

	items := make([]domain.QuestionAndAnswers, 0, len(domainQuestion))

	for _, q := range domainQuestion {
		answerEntities, aErr := s.matchingRepo.GetQuestionAnswers(ctx, q.ID)
		if aErr != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get question answers questionID=%v : %w", q.ID, aErr)
		}

		items = append(items, domain.QuestionAndAnswers{
			Question: q,
			Answers:  mapper.MapAnswerEntitiesToDomain(answerEntities),
		})
	}

	total, err := s.matchingRepo.CountQuestions(ctx, catPtr)
	if err != nil {
		return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get total questions category=%v : %w", category, err)
	}

	return domain.QuestionsAndAnswers{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
