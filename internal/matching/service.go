package matching

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
	"github.com/Haerd-Limited/dating-api/internal/matching/mapper"
	"github.com/Haerd-Limited/dating-api/internal/matching/storage"
)

type Service interface {
	GetQuestionsAndAnswers(ctx context.Context, category string, offset, limit int) (domain.QuestionsAndAnswers, error)
	SaveAnswer(ctx context.Context, cmd domain.SaveAnswerCommand) error
	ComputeMatch(ctx context.Context, viewerID, targetID string, minOverlap int) (*domain.MatchSummary, error)
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

const defaultBadgeLimit = 3

// ComputeMatch calculates the viewer↔target match summary.
// - Uses only overlapping questions (both answered).
// - OkCupid-style: geometric mean of asymmetric satisfactions.
// - Honors mandatory gates in either direction.
// - Hides score when overlap < minOverlap but still returns overlap count.
func (s *service) ComputeMatch(ctx context.Context, viewerID, targetID string, minOverlap int) (*domain.MatchSummary, error) {
	out := &domain.MatchSummary{}

	// trivial self-view: 100% (skip DB, but still try to return overlap count via one call)
	if viewerID == targetID {
		_, _, overlap, err := s.matchingRepo.PerspectiveSums(ctx, viewerID, targetID)
		if err != nil {
			return out, fmt.Errorf("failed to compute perspective sums: %w", err)
		}

		out.MatchPercent = 100
		out.OverlapCount = overlap

		return out, nil
	}

	// 1) Mandatory gate (either side fails → near-zero or hidden)
	mismatchAB, err := s.matchingRepo.HasMandatoryMismatch(ctx, viewerID, targetID)
	if err != nil {
		return out, fmt.Errorf("failed to check mandatory mismatch(AB): %w", err)
	}

	mismatchBA, err := s.matchingRepo.HasMandatoryMismatch(ctx, targetID, viewerID)
	if err != nil {
		return out, fmt.Errorf("failed to check mandatory mismatch(BA): %w", err)
	}

	/*this means someone in this pair, doesn't meet a mandatory requirement of the other user.
	So the relationship is chalked from the get go and no point in calculating match percentage with other overlapping questions*/
	if mismatchAB || mismatchBA {
		// Still compute overlap (for UX messaging), but short-circuit score.
		_, _, overlap, psErr := s.matchingRepo.PerspectiveSums(ctx, viewerID, targetID)
		if psErr != nil {
			return out, fmt.Errorf("failed to compute perspective sums: %w", psErr)
		}

		out.OverlapCount = overlap
		out.MatchPercent = 1
		out.HiddenReason = "Mandatory mismatch"
		out.Badges = []domain.MatchBadge{}

		return out, nil
	}

	// 2) Perspective sums (A→B and B→A)
	earnedAB, totalAB, overlapAB, err := s.matchingRepo.PerspectiveSums(ctx, viewerID, targetID)
	if err != nil {
		return out, fmt.Errorf("failed to compute perspective sums: %w", err)
	}

	earnedBA, totalBA, _, err := s.matchingRepo.PerspectiveSums(ctx, targetID, viewerID)
	if err != nil {
		return out, fmt.Errorf("failed to compute perspective sums: %w", err)
	}

	out.OverlapCount = overlapAB // same either direction by construction

	// 3) Enforce minimum overlap (hide score if not enough)
	if minOverlap > 0 && out.OverlapCount < minOverlap {
		out.HiddenReason = fmt.Sprintf("Not enough overlap (need %d+ shared answers)", minOverlap)
		out.MatchPercent = 0
		out.Badges = []domain.MatchBadge{}

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

	out.MatchPercent = percent

	// 5) Badges (viewer-perspective satisfied, highest-weight)
	badges, err := s.matchingRepo.TopBadges(ctx, viewerID, targetID, defaultBadgeLimit)
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
