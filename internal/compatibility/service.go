package compatibility

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/mapper"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/storage"
)

type Service interface {
	GetOverview(ctx context.Context, userID string) (domain.Overview, error)
	GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int, userID *string, viewAll bool) (domain.QuestionsAndAnswers, error)
	SaveAnswer(ctx context.Context, cmd domain.SaveAnswerCommand) error
	ComputeCompatibility(ctx context.Context, viewerID, targetID string, minOverlap int) (*domain.CompatibilitySummary, error)
}

type service struct {
	logger            *zap.Logger
	compatibilityRepo storage.CompatibilityRepository
}

func NewCompatibilityService(
	logger *zap.Logger,
	compatibilityRepository storage.CompatibilityRepository,
) Service {
	return &service{
		logger:            logger,
		compatibilityRepo: compatibilityRepository,
	}
}

var (
	allowedImportance = map[string]struct{}{
		"irrelevant": {}, "a_little": {}, "somewhat": {}, "very": {}, "mandatory": {},
	}
	ErrInvalidImportance           = fmt.Errorf("invalid importance")
	ErrInvalidAnswerID             = fmt.Errorf("invalid answer_id")
	ErrAcceptableAnswerIDsRequired = fmt.Errorf("acceptable_answer_ids required")
	ErrSequentialAnsweringRequired = fmt.Errorf("questions must be answered sequentially")
)

const defaultBadgeLimit = 3

func (s *service) GetOverview(ctx context.Context, userID string) (domain.Overview, error) {
	categories, err := s.getQuestionCategories(ctx)
	if err != nil {
		return domain.Overview{}, fmt.Errorf("failed to get categories: %w", err)
	}

	questionPacks := make([]domain.Pack, 0, len(categories))

	for _, c := range categories {
		totalCategoryQuestionCount, cErr := s.compatibilityRepo.CountQuestions(ctx, &c.Key)
		if cErr != nil {
			return domain.Overview{}, fmt.Errorf("failed to get count questions category=%v : %w", c.Key, cErr)
		}

		answeredQuestionsCount, cErr := s.compatibilityRepo.CountAnsweredByCategory(ctx, userID, c.Key)
		if cErr != nil {
			return domain.Overview{}, fmt.Errorf("failed to get count answered questions category=%v : %w", c.Key, cErr)
		}

		// Calculate progress percent
		var progressPercent float64
		if totalCategoryQuestionCount > 0 {
			progressPercent = (float64(answeredQuestionsCount) / float64(totalCategoryQuestionCount)) * 100.0
			// Round to 1 decimal place
			progressPercent = math.Round(progressPercent*10) / 10
		}

		pack := domain.Pack{
			CategoryKey:                c.Key,
			CategoryName:               c.Name,
			NumberOfCompletedQuestions: answeredQuestionsCount,
			TotalQuestions:             totalCategoryQuestionCount,
			ProgressPercent:            progressPercent,
		}
		questionPacks = append(questionPacks, pack)
	}

	return domain.Overview{
		QuestionPacks: questionPacks,
	}, nil
}

// ComputeCompatibility calculates the viewer↔target compatibility summary.
// - Uses only overlapping questions (both answered).
// - OkCupid-style: geometric mean of asymmetric satisfactions.
// - Honors mandatory gates in either direction.
// - Hides score when overlap < minOverlap but still returns overlap count.
func (s *service) ComputeCompatibility(ctx context.Context, viewerID, targetID string, minOverlap int) (*domain.CompatibilitySummary, error) {
	out := &domain.CompatibilitySummary{}

	// trivial self-view: 100% (skip DB, but still try to return overlap count via one call)
	if viewerID == targetID {
		_, _, overlap, err := s.compatibilityRepo.PerspectiveSums(ctx, viewerID, targetID)
		if err != nil {
			return out, fmt.Errorf("failed to compute perspective sums: %w", err)
		}

		out.CompatibilityPercent = 100
		out.OverlapCount = overlap

		return out, nil
	}

	// 1) Mandatory gate (either side fails → near-zero or hidden)
	mismatchAB, err := s.compatibilityRepo.HasMandatoryMismatch(ctx, viewerID, targetID)
	if err != nil {
		return out, fmt.Errorf("failed to check mandatory mismatch(AB): %w", err)
	}

	mismatchBA, err := s.compatibilityRepo.HasMandatoryMismatch(ctx, targetID, viewerID)
	if err != nil {
		return out, fmt.Errorf("failed to check mandatory mismatch(BA): %w", err)
	}

	/*this means someone in this pair, doesn't meet a mandatory requirement of the other user.
	So the relationship is chalked from the get go and no point in calculating match percentage with other overlapping questions*/
	if mismatchAB || mismatchBA {
		// Still compute overlap (for UX messaging), but short-circuit score.
		_, _, overlap, psErr := s.compatibilityRepo.PerspectiveSums(ctx, viewerID, targetID)
		if psErr != nil {
			return out, fmt.Errorf("failed to compute perspective sums: %w", psErr)
		}

		out.OverlapCount = overlap
		out.CompatibilityPercent = 1
		out.HiddenReason = "Mandatory mismatch"
		out.Badges = []domain.CompatibilityBadge{}

		return out, nil
	}

	// 2) Perspective sums (A→B and B→A)
	earnedAB, totalAB, overlapAB, err := s.compatibilityRepo.PerspectiveSums(ctx, viewerID, targetID)
	if err != nil {
		return out, fmt.Errorf("failed to compute perspective sums: %w", err)
	}

	earnedBA, totalBA, _, err := s.compatibilityRepo.PerspectiveSums(ctx, targetID, viewerID)
	if err != nil {
		return out, fmt.Errorf("failed to compute perspective sums: %w", err)
	}

	out.OverlapCount = overlapAB // same either direction by construction

	// 3) Enforce minimum overlap (hide score if not enough)
	if minOverlap > 0 && out.OverlapCount < minOverlap {
		out.HiddenReason = fmt.Sprintf("Not enough overlap (need %d+ shared answers)", minOverlap)
		out.CompatibilityPercent = 0
		out.Badges = []domain.CompatibilityBadge{}

		return out, nil
	}

	// 4) Match math (guard divide-by-zero by treating zero-total as denominator=1)
	sA := 0.0
	if totalAB > 0 {
		sA = float64(earnedAB) / float64(totalAB)
	} else {
		sA = 1.0
	}

	sB := 0.0
	if totalBA > 0 {
		sB = float64(earnedBA) / float64(totalBA)
	} else {
		sB = 1.0
	}

	match := math.Sqrt(sA * sB)

	percent := int(math.Round(match * 100.0))
	if percent < 0 {
		percent = 0
	}

	if percent > 100 {
		percent = 100
	}

	out.CompatibilityPercent = percent

	// 5) Badges (viewer-perspective satisfied, highest-weight)
	badges, err := s.compatibilityRepo.TopBadges(ctx, viewerID, targetID, defaultBadgeLimit)
	if err != nil {
		return out, fmt.Errorf("failed to get top badges: %w", err)
	}

	out.Badges = badges

	return out, nil
}

func (s *service) SaveAnswer(ctx context.Context, cmd domain.SaveAnswerCommand) error {
	// 1) validate importance
	if _, ok := allowedImportance[cmd.Importance]; !ok {
		return fmt.Errorf("%w: %s", ErrInvalidImportance, cmd.Importance)
	}

	// 2) load all valid answers for this question
	valid, err := s.compatibilityRepo.GetAnswerIDsForQuestion(ctx, cmd.QuestionID)
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

	// 6) Validate sequential answering (unless updating an existing answer)
	existingAnswer, err := s.compatibilityRepo.GetUserAnswerForQuestion(ctx, cmd.UserID, cmd.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to check existing answer: %w", err)
	}

	// If this is a new answer (not an update), validate sequential order
	if existingAnswer == nil {
		if err := s.validateSequentialAnswering(ctx, cmd.UserID, cmd.QuestionID); err != nil {
			return err
		}
	}

	// 7) persist
	return s.compatibilityRepo.UpsertUserAnswer(ctx, mapper.MapSaveAnswerCommandToUserAnswerEntity(cmd))
}

func (s *service) GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int, userID *string, viewAll bool) (domain.QuestionsAndAnswers, error) {
	// If userID is provided and viewAll is false, return only the next unanswered question
	if userID != nil && category != "" && !viewAll {
		nextQuestion, err := s.compatibilityRepo.GetNextUnansweredQuestion(ctx, *userID, category)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get next unanswered question: %w", err)
		}

		// If no unanswered questions, return empty result
		if nextQuestion == nil {
			total, err := s.compatibilityRepo.CountQuestions(ctx, &category)
			if err != nil {
				return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get total questions: %w", err)
			}

			return domain.QuestionsAndAnswers{
				Items:  []domain.QuestionAndAnswers{},
				Total:  total,
				Limit:  limit,
				Offset: 0,
			}, nil
		}

		// Get question answers
		answerEntities, err := s.compatibilityRepo.GetQuestionAnswers(ctx, nextQuestion.ID)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get question answers: %w", err)
		}

		// Check if user has an existing answer (for update flow)
		var userAnswer *domain.UserAnswer

		existingAnswer, err := s.compatibilityRepo.GetUserAnswerForQuestion(ctx, *userID, nextQuestion.ID)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get existing answer: %w", err)
		}

		if existingAnswer != nil {
			userAnswer = mapper.MapUserAnswerEntityToDomain(existingAnswer)
		}

		domainQuestion := mapper.MapQuestionEntityToDomain(nextQuestion)
		items := []domain.QuestionAndAnswers{
			{
				Question:   domainQuestion,
				Answers:    mapper.MapAnswerEntitiesToDomain(answerEntities),
				UserAnswer: userAnswer,
			},
		}

		total, err := s.compatibilityRepo.CountQuestions(ctx, &category)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get total questions: %w", err)
		}

		return domain.QuestionsAndAnswers{
			Items:  items,
			Total:  total,
			Limit:  1,
			Offset: 0,
		}, nil
	}

	// If viewAll is true and userID is provided, return all questions in the category with their answered status
	if viewAll && userID != nil && category != "" {
		// Get all questions in order for this category
		allQuestionEntities, err := s.compatibilityRepo.GetQuestionsInOrder(ctx, category)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get questions in order: %w", err)
		}

		// Get category info
		categories, err := s.getQuestionCategories(ctx)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get categories: %w", err)
		}

		var categoryName string

		for _, c := range categories {
			if c.Key == category {
				categoryName = c.Name
				break
			}
		}

		items := make([]domain.QuestionAndAnswers, 0, len(allQuestionEntities))
		answeredCount := 0

		var nextQuestionID *int64

		for _, qe := range allQuestionEntities {
			// Get question answers
			answerEntities, err := s.compatibilityRepo.GetQuestionAnswers(ctx, qe.ID)
			if err != nil {
				return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get question answers: %w", err)
			}

			// Check for existing answer
			var userAnswer *domain.UserAnswer

			existingAnswer, err := s.compatibilityRepo.GetUserAnswerForQuestion(ctx, *userID, qe.ID)
			if err != nil {
				return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get existing answer: %w", err)
			}

			if existingAnswer != nil {
				userAnswer = mapper.MapUserAnswerEntityToDomain(existingAnswer)
				answeredCount++
			} else if nextQuestionID == nil {
				// First unanswered question
				nextQuestionID = &qe.ID
			}

			domainQuestion := mapper.MapQuestionEntityToDomain(qe)
			items = append(items, domain.QuestionAndAnswers{
				Question:   domainQuestion,
				Answers:    mapper.MapAnswerEntitiesToDomain(answerEntities),
				UserAnswer: userAnswer,
			})
		}

		total, err := s.compatibilityRepo.CountQuestions(ctx, &category)
		if err != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get total questions: %w", err)
		}

		// Calculate progress percent
		var progressPercent float64
		if total > 0 {
			progressPercent = (float64(answeredCount) / float64(total)) * 100.0
			progressPercent = math.Round(progressPercent*10) / 10
		}

		return domain.QuestionsAndAnswers{
			Items:  items,
			Total:  total,
			Limit:  total, // Return all, so limit equals total
			Offset: 0,
			ProgressSummary: &domain.ProgressSummary{
				CategoryKey:                category,
				CategoryName:               categoryName,
				NumberOfCompletedQuestions: answeredCount,
				TotalQuestions:             total,
				ProgressPercent:            progressPercent,
				NextQuestionID:             nextQuestionID,
			},
		}, nil
	}

	// Backward compatibility: return questions in order with pagination
	domainQuestions, err := s.listQuestions(ctx, category, offset, limit)
	if err != nil {
		return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to list questions category=%v : %w", category, err)
	}

	items := make([]domain.QuestionAndAnswers, 0, len(domainQuestions))

	for _, q := range domainQuestions {
		answerEntities, aErr := s.compatibilityRepo.GetQuestionAnswers(ctx, q.ID)
		if aErr != nil {
			return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get question answers questionID=%v : %w", q.ID, aErr)
		}

		// If userID provided, check for existing answer
		var userAnswer *domain.UserAnswer

		if userID != nil {
			existingAnswer, err := s.compatibilityRepo.GetUserAnswerForQuestion(ctx, *userID, q.ID)
			if err != nil {
				return domain.QuestionsAndAnswers{}, fmt.Errorf("failed to get existing answer: %w", err)
			}

			if existingAnswer != nil {
				userAnswer = mapper.MapUserAnswerEntityToDomain(existingAnswer)
			}
		}

		items = append(items, domain.QuestionAndAnswers{
			Question:   q,
			Answers:    mapper.MapAnswerEntitiesToDomain(answerEntities),
			UserAnswer: userAnswer,
		})
	}

	total, err := s.compatibilityRepo.CountQuestions(ctx, &category)
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

func (s *service) listQuestions(ctx context.Context, category string, offset, limit int) ([]domain.Question, error) {
	var catPtr *string
	if category != "" {
		catPtr = &category
	}

	questionEntities, err := s.compatibilityRepo.ListQuestions(ctx, catPtr, limit, offset)
	if err != nil {
		return nil, err
	}

	return mapper.MapQuestionEntitiesToDomain(questionEntities), nil
}

func (s *service) getQuestionCategories(ctx context.Context) ([]domain.QuestionCategory, error) {
	categoryEntities, err := s.compatibilityRepo.GetQuestionCategories(ctx)
	if err != nil {
		return nil, err
	}

	var categories []domain.QuestionCategory
	for _, c := range categoryEntities {
		categories = append(categories, domain.QuestionCategory{
			ID:        c.ID,
			Name:      c.Name,
			Key:       c.Key,
			CreatedAt: c.CreatedAt,
		})
	}

	return categories, nil
}

// validateSequentialAnswering ensures that all previous questions in the category have been answered
// before allowing the user to answer the current question.
func (s *service) validateSequentialAnswering(ctx context.Context, userID string, questionID int64) error {
	// Get the question's sort_order and category
	_, categoryKey, err := s.compatibilityRepo.GetQuestionSortOrderAndCategory(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get question info: %w", err)
	}

	// Get all questions in order for this category
	allQuestions, err := s.compatibilityRepo.GetQuestionsInOrder(ctx, categoryKey)
	if err != nil {
		return fmt.Errorf("failed to get questions in order: %w", err)
	}

	// Find the target question's position (index) in the ordered list
	var targetIndex int = -1

	for i, q := range allQuestions {
		if q.ID == questionID {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("question %d not found in category %s", questionID, categoryKey)
	}

	// Get all answered question IDs for this user in this category
	answeredIDs, err := s.compatibilityRepo.GetUserAnsweredQuestionIDs(ctx, userID, categoryKey)
	if err != nil {
		return fmt.Errorf("failed to get answered questions: %w", err)
	}

	answeredSet := make(map[int64]struct{}, len(answeredIDs))
	for _, id := range answeredIDs {
		answeredSet[id] = struct{}{}
	}

	// Check if all questions before this one (by index) have been answered
	for i := 0; i < targetIndex; i++ {
		q := allQuestions[i]
		if _, answered := answeredSet[q.ID]; !answered {
			return fmt.Errorf("%w: previous questions in category must be answered before question %d",
				ErrSequentialAnsweringRequired, questionID)
		}
	}

	return nil
}
