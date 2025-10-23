package matching

import (
	"context"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
	"github.com/Haerd-Limited/dating-api/internal/matching/mapper"
	"github.com/Haerd-Limited/dating-api/internal/matching/storage"
)

type Service interface {
	GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int) (domain.QuestionsAndAnswers, error)
	SaveAnswer(ctx context.Context, cmd domain.SaveAnswerCommand) error
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

var (
	allowedImportance = map[string]struct{}{
		"irrelevant": {}, "a_little": {}, "somewhat": {}, "very": {}, "mandatory": {},
	}
	ErrInvalidImportance           = fmt.Errorf("invalid importance")
	ErrInvalidAnswerID             = fmt.Errorf("invalid answer_id")
	ErrAcceptableAnswerIDsRequired = fmt.Errorf("acceptable_answer_ids required")
)

func (s *service) SaveAnswer(ctx context.Context, cmd domain.SaveAnswerCommand) error {
	// 1) validate importance
	if _, ok := allowedImportance[cmd.Importance]; !ok {
		return fmt.Errorf("%w: %s", ErrInvalidImportance, cmd.Importance)
	}

	// 2) load all valid answers for this question
	valid, err := s.matchingRepo.GetAnswerIDsForQuestion(ctx, cmd.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to get answer IDs for question_id=%d: %w", cmd.QuestionID, err)
	}

	if len(valid) == 0 {
		return fmt.Errorf("%w : unknown question_id: %d", ErrInvalidAnswerID, cmd.QuestionID)
	}

	validSet := make(map[int64]struct{}, len(valid))
	for _, id := range valid {
		validSet[id] = struct{}{}
	}

	// 3) verify answer_id belongs to question
	if _, ok := validSet[cmd.AnswerID]; !ok {
		return fmt.Errorf("%w : answer_id %d does not belong to question_id %d", ErrInvalidAnswerID, cmd.AnswerID, cmd.QuestionID)
	}

	// 4) clean acceptable ids (dedupe + subset)
	clean := make([]int64, 0, len(cmd.AcceptableAnswerIDs))
	seen := make(map[int64]struct{}, len(cmd.AcceptableAnswerIDs))

	for _, id := range cmd.AcceptableAnswerIDs {
		if _, ok := validSet[id]; !ok {
			continue
		} // drop foreign ids

		if _, dup := seen[id]; dup {
			continue
		} // drop dups

		seen[id] = struct{}{}

		clean = append(clean, id)
	}

	sort.Slice(clean, func(i, j int) bool { return clean[i] < clean[j] })

	// 5) rule: if importance != irrelevant → acceptable must not be empty
	if cmd.Importance != "irrelevant" && len(clean) == 0 {
		return fmt.Errorf("%w when importance is %q", ErrAcceptableAnswerIDsRequired, cmd.Importance)
	}

	// 6) persist
	return s.matchingRepo.UpsertUserAnswer(ctx, mapper.MapSaveAnswerCommandToUserAnswerEntity(cmd))
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
