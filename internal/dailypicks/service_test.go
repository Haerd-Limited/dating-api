package dailypicks

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/dailypicks/domain"
	dailypicksstorage "github.com/Haerd-Limited/dating-api/internal/dailypicks/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
)

type fakeTx struct {
	committed bool
}

func (t *fakeTx) Commit() error {
	t.committed = true
	return nil
}

func (t *fakeTx) Rollback() error { return nil }

func (t *fakeTx) Raw() *sql.Tx { return nil }

type fakeUoW struct {
	tx *fakeTx
}

func (u *fakeUoW) Begin(_ context.Context) (uow.Tx, error) {
	return u.tx, nil
}

func newTestService(t *testing.T, repo dailypicksstorage.Repository) *service {
	t.Helper()

	return &service{
		logger: zaptest.NewLogger(t),
		repo:   repo,
		uow:    &fakeUoW{tx: &fakeTx{}},
	}
}

func TestSelectValidPicksPreservesRankingOrder(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)

	candidates := []domain.ScoredCandidate{
		{UserID: "high-prefs", PrefsSatisfied: 3, CompatibilityPercent: 60},
		{UserID: "low-prefs", PrefsSatisfied: 2, CompatibilityPercent: 95},
	}

	repo.EXPECT().IsTargetStillValid(ctx, "viewer", "high-prefs").Return(true, nil)
	repo.EXPECT().IsTargetStillValid(ctx, "viewer", "low-prefs").Return(true, nil)

	svc := newTestService(t, repo)
	picks := svc.selectValidPicks(ctx, "viewer", candidates, 2)

	require.Len(t, picks, 2)
	assert.Equal(t, "high-prefs", picks[0].TargetUserID)
	assert.Equal(t, "low-prefs", picks[1].TargetUserID)
}

func TestResolveVisibleBatchGatedMatchLimit(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)

	gomock.InOrder(
		repo.EXPECT().LockUserForPicks(ctx, gomock.Nil(), "user-1").Return(nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchRevealed).Return(&domain.Batch{
			ID: "batch-1",
		}, nil),
		repo.EXPECT().CountUndecidedPicks(ctx, gomock.Nil(), "batch-1").Return(1, nil),
		repo.EXPECT().CountActiveMatches(ctx, gomock.Nil(), "user-1").Return(int64(1), nil),
	)

	svc := newTestService(t, repo)
	res, err := svc.resolveVisibleBatch(ctx, nil, "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StateGatedMatchLimit, res.State)
	assert.Nil(t, res.Batch)
}

func TestResolveVisibleBatchPromotesPending(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)

	now := time.Now().UTC()
	pending := &domain.Batch{
		ID:     "pending-1",
		UserID: "user-1",
		State:  domain.BatchPending,
		Picks: []domain.Pick{
			{ID: "pick-1", Status: domain.PickPending, TargetUserID: "target-1"},
			{ID: "pick-2", Status: domain.PickPending, TargetUserID: "target-2"},
		},
	}

	gomock.InOrder(
		repo.EXPECT().LockUserForPicks(ctx, gomock.Nil(), "user-1").Return(nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchRevealed).Return(nil, nil),
		repo.EXPECT().CountActiveMatches(ctx, gomock.Nil(), "user-1").Return(int64(0), nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchPending).Return(pending, nil),
		repo.EXPECT().MarkBatchState(ctx, gomock.Nil(), "pending-1", domain.BatchRevealed, gomock.Any()).Return(nil),
		repo.EXPECT().IsTargetStillValid(ctx, "user-1", "target-1").Return(true, nil),
		repo.EXPECT().IsTargetStillValid(ctx, "user-1", "target-2").Return(true, nil),
		repo.EXPECT().GetBatchByID(ctx, gomock.Nil(), "pending-1").Return(pending, nil),
	)

	svc := newTestService(t, repo)
	res, err := svc.resolveVisibleBatch(ctx, nil, "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StateRevealed, res.State)
	assert.True(t, res.NewlyRevealed)
	require.NotNil(t, res.Batch)
	assert.Equal(t, domain.BatchRevealed, res.Batch.State)

	_ = now
}

func TestResolveVisibleBatchAwaitingNext(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)

	gomock.InOrder(
		repo.EXPECT().LockUserForPicks(ctx, gomock.Nil(), "user-1").Return(nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchRevealed).Return(nil, nil),
		repo.EXPECT().CountActiveMatches(ctx, gomock.Nil(), "user-1").Return(int64(0), nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchPending).Return(nil, nil),
		repo.EXPECT().HasAnyBatch(ctx, gomock.Nil(), "user-1").Return(true, nil),
	)

	svc := newTestService(t, repo)
	res, err := svc.resolveVisibleBatch(ctx, nil, "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StateAwaitingNext, res.State)
}

func TestGenerateForUserSkipsEmptySeekGender(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)

	repo.EXPECT().SeekGenderIDs(ctx, "user-1").Return(nil, nil)

	svc := newTestService(t, repo)
	id, err := svc.generateForUser(ctx, nil, "user-1")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestMarkDecidedAcquiresLockBeforePromotion(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := dailypicksstorage.NewMockRepository(ctrl)
	tx := &fakeTx{}

	gomock.InOrder(
		repo.EXPECT().LockUserForPicks(ctx, gomock.Nil(), "user-1").Return(nil),
		repo.EXPECT().MarkPickDecided(ctx, gomock.Nil(), "user-1", "target-1", domain.PickLiked).Return(nil),
		repo.EXPECT().GetBatchByState(ctx, gomock.Nil(), "user-1", domain.BatchRevealed).Return(&domain.Batch{ID: "batch-1"}, nil),
		repo.EXPECT().CountUndecidedPicks(ctx, gomock.Nil(), "batch-1").Return(1, nil),
	)

	svc := &service{
		logger: zaptest.NewLogger(t),
		repo:   repo,
		uow:    &fakeUoW{tx: tx},
	}

	err := svc.MarkDecided(ctx, nil, "user-1", "target-1", "like")
	require.NoError(t, err)
	assert.True(t, tx.committed)
}
