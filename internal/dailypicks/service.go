package dailypicks

//go:generate mockgen -source=service.go -destination=service_mock.go -package=dailypicks

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/dailypicks/domain"
	dailypicksstorage "github.com/Haerd-Limited/dating-api/internal/dailypicks/storage"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

const (
	shortlistSize     = 10
	picksPerBatch     = 2
	generationWorkers = 8
)

type Service interface {
	GetDailyPicks(ctx context.Context, userID string) (domain.DailyPicksResult, error)
	MarkDecided(ctx context.Context, exec boil.ContextExecutor, userID, targetUserID, action string) error
	GenerateForAllUsers(ctx context.Context) (domain.GenerationStats, error)
}

type service struct {
	repo                dailypicksstorage.Repository
	profileService      profile.Service
	notificationService notification.Service
	uow                 uow.UoW
	logger              *zap.Logger
}

func NewService(
	logger *zap.Logger,
	repo dailypicksstorage.Repository,
	profileService profile.Service,
	notificationService notification.Service,
	unitOfWork uow.UoW,
) Service {
	return &service{
		repo:                repo,
		profileService:      profileService,
		notificationService: notificationService,
		uow:                 unitOfWork,
		logger:              logger,
	}
}

func (s *service) GetDailyPicks(ctx context.Context, userID string) (domain.DailyPicksResult, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return domain.DailyPicksResult{}, fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	resolved, err := s.resolveVisibleBatch(ctx, tx.Raw(), userID)
	if err != nil {
		return domain.DailyPicksResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.DailyPicksResult{}, fmt.Errorf("commit tx: %w", err)
	}

	activeMatches, err := s.repo.CountActiveMatches(ctx, nil, userID)
	if err != nil {
		return domain.DailyPicksResult{}, fmt.Errorf("count active matches: %w", err)
	}

	result := domain.DailyPicksResult{
		State:              resolved.State,
		ActiveMatchesCount: activeMatches,
		MatchSlotLimit:     domain.MatchSlotLimit,
		NextDropAt:         nextDropAt(time.Now()),
	}

	if resolved.Batch != nil {
		cards, undecided, err := s.hydratePicks(ctx, userID, resolved.Batch)
		if err != nil {
			return domain.DailyPicksResult{}, err
		}

		result.Picks = cards
		result.UndecidedCount = undecided
	} else {
		result.Picks = []profilecard.ProfileCard{}
	}

	return result, nil
}

func (s *service) MarkDecided(ctx context.Context, exec boil.ContextExecutor, userID, targetUserID, action string) error {
	status := mapSwipeActionToPickStatus(action)
	if status == "" {
		return nil
	}

	if exec != nil {
		if err := s.repo.MarkPickDecided(ctx, exec, userID, targetUserID, status); err != nil {
			return err
		}

		return nil
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	if err := s.repo.LockUserForPicks(ctx, tx.Raw(), userID); err != nil {
		return err
	}

	if err := s.repo.MarkPickDecided(ctx, tx.Raw(), userID, targetUserID, status); err != nil {
		return err
	}

	if err := s.completeAndPromoteIfNeeded(ctx, tx.Raw(), userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) GenerateForAllUsers(ctx context.Context) (domain.GenerationStats, error) {
	userIDs, err := s.repo.ListEligibleUserIDs(ctx)
	if err != nil {
		return domain.GenerationStats{}, err
	}

	stats := domain.GenerationStats{UsersProcessed: len(userIDs)}

	var notifyMu sync.Mutex

	var newlyRevealed []string

	var notifyBatchIDs []string

	jobs := make(chan string, len(userIDs))
	for _, id := range userIDs {
		jobs <- id
	}

	close(jobs)

	workerCount := generationWorkers
	if workerCount > len(userIDs) {
		workerCount = len(userIDs)
	}

	if workerCount == 0 {
		return stats, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for userID := range jobs {
				if ctx.Err() != nil {
					return
				}

				created, batchID, revealed, err := s.generateAndResolveUser(ctx, userID)
				if err != nil {
					s.logger.Sugar().Warnw("daily picks user generation failed", "error", err, "userID", userID)
					continue
				}

				notifyMu.Lock()
				if created {
					stats.BatchesCreated++
				} else {
					stats.Skipped++
				}

				if revealed {
					stats.Notified++

					newlyRevealed = append(newlyRevealed, userID)

					if batchID != "" {
						notifyBatchIDs = append(notifyBatchIDs, batchID)
					}
				}
				notifyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(newlyRevealed) > 0 {
		if err := s.notificationService.SendDailyPicksNotification(ctx, newlyRevealed); err != nil {
			s.logger.Sugar().Warnw("send daily picks notification", "error", err, "count", len(newlyRevealed))
		} else if err := s.repo.MarkBatchNotified(ctx, notifyBatchIDs); err != nil {
			s.logger.Sugar().Warnw("mark batches notified", "error", err)
		}
	}

	return stats, nil
}

func (s *service) generateAndResolveUser(ctx context.Context, userID string) (batchCreated bool, batchID string, newlyRevealed bool, err error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return false, "", false, err
	}

	defer func() { _ = tx.Rollback() }()

	if err := s.repo.LockUserForPicks(ctx, tx.Raw(), userID); err != nil {
		return false, "", false, err
	}

	createdID, err := s.generateForUser(ctx, tx.Raw(), userID)
	if err != nil {
		return false, "", false, err
	}

	resolved, err := s.resolveVisibleBatch(ctx, tx.Raw(), userID)
	if err != nil {
		return createdID != "", createdID, false, err
	}

	if err := tx.Commit(); err != nil {
		return createdID != "", createdID, false, err
	}

	if resolved.NewlyRevealed && resolved.Batch != nil {
		return createdID != "", resolved.Batch.ID, true, nil
	}

	return createdID != "", createdID, false, nil
}

func (s *service) resolveVisibleBatch(ctx context.Context, sqlTx *sql.Tx, userID string) (domain.ResolveResult, error) {
	if err := s.repo.LockUserForPicks(ctx, sqlTx, userID); err != nil {
		return domain.ResolveResult{}, err
	}

	revealed, err := s.repo.GetBatchByState(ctx, sqlTx, userID, domain.BatchRevealed)
	if err != nil {
		return domain.ResolveResult{}, err
	}

	if revealed != nil {
		undecided, err := s.repo.CountUndecidedPicks(ctx, sqlTx, revealed.ID)
		if err != nil {
			return domain.ResolveResult{}, err
		}

		if undecided == 0 {
			if err := s.repo.MarkBatchState(ctx, sqlTx, revealed.ID, domain.BatchCompleted, nil); err != nil {
				return domain.ResolveResult{}, err
			}

			revealed = nil
		}
	}

	activeMatches, err := s.repo.CountActiveMatches(ctx, sqlTx, userID)
	if err != nil {
		return domain.ResolveResult{}, err
	}

	if activeMatches > 0 {
		return domain.ResolveResult{State: domain.StateGatedMatchLimit}, nil
	}

	if revealed != nil {
		revealed, err = s.validateAndBackfillPicks(ctx, sqlTx, userID, revealed)
		if err != nil {
			return domain.ResolveResult{}, err
		}

		if revealed != nil && s.batchHasPendingPicks(revealed) {
			return domain.ResolveResult{State: domain.StateRevealed, Batch: revealed}, nil
		}
	}

	pending, err := s.repo.GetBatchByState(ctx, sqlTx, userID, domain.BatchPending)
	if err != nil {
		return domain.ResolveResult{}, err
	}

	if pending != nil {
		now := time.Now().UTC()
		if err := s.repo.MarkBatchState(ctx, sqlTx, pending.ID, domain.BatchRevealed, &now); err != nil {
			return domain.ResolveResult{}, err
		}

		pending.State = domain.BatchRevealed
		pending.RevealedAt = &now

		pending, err = s.validateAndBackfillPicks(ctx, sqlTx, userID, pending)
		if err != nil {
			return domain.ResolveResult{}, err
		}

		if pending == nil || !s.batchHasPendingPicks(pending) {
			return domain.ResolveResult{State: domain.StateNone}, nil
		}

		return domain.ResolveResult{
			State:         domain.StateRevealed,
			Batch:         pending,
			NewlyRevealed: true,
		}, nil
	}

	hasBatch, err := s.repo.HasAnyBatch(ctx, sqlTx, userID)
	if err != nil {
		return domain.ResolveResult{}, err
	}

	if !hasBatch {
		createdID, genErr := s.generateForUser(ctx, sqlTx, userID)
		if genErr != nil {
			s.logger.Sugar().Warnw("cold start generation failed", "error", genErr, "userID", userID)
			return domain.ResolveResult{State: domain.StateAwaitingNext}, nil
		}

		if createdID == "" {
			return domain.ResolveResult{State: domain.StateNone}, nil
		}

		pending, err = s.repo.GetBatchByState(ctx, sqlTx, userID, domain.BatchPending)
		if err != nil {
			return domain.ResolveResult{}, err
		}

		if pending == nil {
			return domain.ResolveResult{State: domain.StateAwaitingNext}, nil
		}

		now := time.Now().UTC()
		if err := s.repo.MarkBatchState(ctx, sqlTx, pending.ID, domain.BatchRevealed, &now); err != nil {
			return domain.ResolveResult{}, err
		}

		pending.State = domain.BatchRevealed
		pending.RevealedAt = &now

		pending, err = s.validateAndBackfillPicks(ctx, sqlTx, userID, pending)
		if err != nil {
			return domain.ResolveResult{}, err
		}

		if pending == nil || !s.batchHasPendingPicks(pending) {
			return domain.ResolveResult{State: domain.StateNone}, nil
		}

		return domain.ResolveResult{State: domain.StateRevealed, Batch: pending}, nil
	}

	return domain.ResolveResult{State: domain.StateAwaitingNext}, nil
}

func (s *service) completeAndPromoteIfNeeded(ctx context.Context, sqlTx *sql.Tx, userID string) error {
	revealed, err := s.repo.GetBatchByState(ctx, sqlTx, userID, domain.BatchRevealed)
	if err != nil {
		return err
	}

	if revealed == nil {
		return nil
	}

	undecided, err := s.repo.CountUndecidedPicks(ctx, sqlTx, revealed.ID)
	if err != nil {
		return err
	}

	if undecided > 0 {
		return nil
	}

	if err := s.repo.MarkBatchState(ctx, sqlTx, revealed.ID, domain.BatchCompleted, nil); err != nil {
		return err
	}

	activeMatches, err := s.repo.CountActiveMatches(ctx, sqlTx, userID)
	if err != nil {
		return err
	}

	if activeMatches > 0 {
		return nil
	}

	pending, err := s.repo.GetBatchByState(ctx, sqlTx, userID, domain.BatchPending)
	if err != nil {
		return err
	}

	if pending == nil {
		return nil
	}

	now := time.Now().UTC()
	if err := s.repo.MarkBatchState(ctx, sqlTx, pending.ID, domain.BatchRevealed, &now); err != nil {
		return err
	}

	return nil
}

func (s *service) generateForUser(ctx context.Context, sqlTx *sql.Tx, userID string) (string, error) {
	seekGenderIDs, err := s.repo.SeekGenderIDs(ctx, userID)
	if err != nil {
		return "", err
	}

	if len(seekGenderIDs) == 0 {
		return "", nil
	}

	candidates, err := s.repo.ScoreCandidates(ctx, userID, seekGenderIDs, shortlistSize)
	if err != nil {
		return "", err
	}

	picks := s.selectValidPicks(ctx, userID, candidates, picksPerBatch)
	if len(picks) == 0 {
		return "", nil
	}

	batchDate := londonBatchDate(time.Now())

	return s.repo.ReplacePendingBatch(ctx, sqlTx, userID, batchDate, picks)
}

func (s *service) selectValidPicks(ctx context.Context, userID string, candidates []domain.ScoredCandidate, limit int) []domain.Pick {
	picks := make([]domain.Pick, 0, limit)
	position := 1

	for _, c := range candidates {
		if len(picks) >= limit {
			break
		}

		valid, err := s.repo.IsTargetStillValid(ctx, userID, c.UserID)
		if err != nil || !valid {
			continue
		}

		picks = append(picks, domain.Pick{
			UserID:               userID,
			TargetUserID:         c.UserID,
			Position:             position,
			CompatibilityPercent: c.CompatibilityPercent,
			OverlapCount:         c.OverlapCount,
			PrefsSatisfied:       c.PrefsSatisfied,
			Status:               domain.PickPending,
		})
		position++
	}

	return picks
}

func (s *service) validateAndBackfillPicks(ctx context.Context, sqlTx *sql.Tx, userID string, batch *domain.Batch) (*domain.Batch, error) {
	var expiredIDs []string

	pendingCount := 0

	for _, p := range batch.Picks {
		if p.Status != domain.PickPending {
			continue
		}

		valid, err := s.repo.IsTargetStillValid(ctx, userID, p.TargetUserID)
		if err != nil {
			return nil, err
		}

		if !valid {
			expiredIDs = append(expiredIDs, p.ID)
			continue
		}

		pendingCount++
	}

	if len(expiredIDs) > 0 {
		if err := s.repo.ExpirePicks(ctx, sqlTx, expiredIDs); err != nil {
			return nil, err
		}
	}

	if pendingCount >= picksPerBatch {
		return s.repo.GetBatchByID(ctx, sqlTx, batch.ID)
	}

	needed := picksPerBatch - pendingCount
	if needed <= 0 {
		return s.repo.GetBatchByID(ctx, sqlTx, batch.ID)
	}

	seekGenderIDs, err := s.repo.SeekGenderIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(seekGenderIDs) == 0 {
		return s.repo.GetBatchByID(ctx, sqlTx, batch.ID)
	}

	exclude := make(map[string]struct{})

	for _, p := range batch.Picks {
		if p.Status != domain.PickExpired {
			exclude[p.TargetUserID] = struct{}{}
		}
	}

	candidates, err := s.repo.ScoreCandidates(ctx, userID, seekGenderIDs, shortlistSize)
	if err != nil {
		return nil, err
	}

	backfill := make([]domain.Pick, 0, needed)
	nextPosition := len(batch.Picks) + 1

	for _, c := range candidates {
		if len(backfill) >= needed {
			break
		}

		if _, skip := exclude[c.UserID]; skip {
			continue
		}

		valid, err := s.repo.IsTargetStillValid(ctx, userID, c.UserID)
		if err != nil || !valid {
			continue
		}

		backfill = append(backfill, domain.Pick{
			UserID:               userID,
			TargetUserID:         c.UserID,
			Position:             nextPosition,
			CompatibilityPercent: c.CompatibilityPercent,
			OverlapCount:         c.OverlapCount,
			PrefsSatisfied:       c.PrefsSatisfied,
			Status:               domain.PickPending,
		})
		nextPosition++
	}

	if len(backfill) > 0 {
		if err := s.repo.InsertPicks(ctx, sqlTx, batch.ID, userID, backfill); err != nil {
			return nil, err
		}
	}

	return s.repo.GetBatchByID(ctx, sqlTx, batch.ID)
}

func (s *service) batchHasPendingPicks(batch *domain.Batch) bool {
	for _, p := range batch.Picks {
		if p.Status == domain.PickPending {
			return true
		}
	}

	return false
}

func (s *service) hydratePicks(ctx context.Context, userID string, batch *domain.Batch) ([]profilecard.ProfileCard, int, error) {
	viewer, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("get viewer profile: %w", err)
	}

	cards := make([]profilecard.ProfileCard, 0, len(batch.Picks))
	undecided := 0

	for _, p := range batch.Picks {
		if p.Status != domain.PickPending {
			continue
		}

		undecided++

		card, err := s.profileService.GetProfileCardWithDistance(ctx, p.TargetUserID, viewer.Latitude, viewer.Longitude)
		if err != nil {
			return nil, 0, fmt.Errorf("hydrate pick target=%s: %w", p.TargetUserID, err)
		}

		card.CompatibilitySummary = &profilecard.CompatibilitySummary{
			CompatibilityPercent: p.CompatibilityPercent,
			OverlapCount:         p.OverlapCount,
		}
		cards = append(cards, card)
	}

	return cards, undecided, nil
}

func mapSwipeActionToPickStatus(action string) string {
	switch action {
	case constants.ActionLike, constants.ActionSuperlike:
		return domain.PickLiked
	case constants.ActionPass:
		return domain.PickPassed
	default:
		return ""
	}
}

func londonBatchDate(now time.Time) time.Time {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		loc = time.UTC
	}

	t := now.In(loc)

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func nextDropAt(now time.Time) *time.Time {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		loc = time.UTC
	}

	local := now.In(loc)
	next := time.Date(local.Year(), local.Month(), local.Day(), 19, 0, 0, 0, loc)

	if !next.After(local) {
		next = next.Add(24 * time.Hour)
	}

	utc := next.UTC()

	return &utc
}
