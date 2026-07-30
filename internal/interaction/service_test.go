package interaction

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	dailypicks "github.com/Haerd-Limited/dating-api/internal/dailypicks"
	discoverdomain "github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	interactiondomain "github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

// newServiceWithRepo builds a partially-wired service exposing only the
// fields each test needs. Other dependencies are intentionally left nil and
// any method that would touch them is out of scope for these tests.
func newServiceWithRepo(t *testing.T, repo storage.InteractionRepository) *service {
	t.Helper()

	return &service{
		logger:          zaptest.NewLogger(t),
		interactionRepo: repo,
	}
}

type fakeTx struct {
	committed bool
}

func (t *fakeTx) Commit() error {
	t.committed = true
	return nil
}

func (t *fakeTx) Rollback() error {
	return nil
}

func (t *fakeTx) Raw() *sql.Tx {
	return nil
}

type fakeUoW struct {
	tx *fakeTx
}

func (u *fakeUoW) Begin(_ context.Context) (uow.Tx, error) {
	return u.tx, nil
}

type fakeDiscoverService struct {
	alreadyInteracted bool
}

func (s fakeDiscoverService) GetDiscoverFeed(context.Context, string, int, int) (discoverdomain.DiscoverFeedResult, error) {
	return discoverdomain.DiscoverFeedResult{}, nil
}

func (s fakeDiscoverService) GetDiscoverFeedWithFilters(context.Context, string, int, int, *discoverdomain.DiscoverFilters) (discoverdomain.DiscoverFeedResult, error) {
	return discoverdomain.DiscoverFeedResult{}, nil
}

func (s fakeDiscoverService) GetVoiceWorthHearing(context.Context, string) ([]profilecard.ProfileCard, error) {
	return nil, nil
}

func (s fakeDiscoverService) GetVoiceWorthHearingIDs(context.Context, string) ([]string, error) {
	return nil, nil
}

func (s fakeDiscoverService) AlreadyInteracted(context.Context, string, string) (bool, error) {
	return s.alreadyInteracted, nil
}

func (s fakeDiscoverService) GetUserPreferences(context.Context, string) (*discoverdomain.StoredDiscoverPreferences, error) {
	return nil, nil
}

func (s fakeDiscoverService) ComputeCompatibility(context.Context, string, string) (*profilecard.CompatibilitySummary, error) {
	return nil, nil
}

type fakeBroadcaster struct{}

func (fakeBroadcaster) BroadcastToConversation(string, []byte) {}

func (fakeBroadcaster) BroadcastToUser(string, []byte) {}

// TestEnforceActiveMatchCap covers the three branches of the symmetric cap:
// actor full, target full, both under. The advisory-lock acquisition is the
// caller's responsibility and is verified by code review + Postgres semantics
// (see the plan's Decisions section).
func TestEnforceActiveMatchCap(t *testing.T) {
	const (
		actorID  = "actor-1"
		targetID = "target-2"
	)

	ctx := context.Background()

	cases := []struct {
		name      string
		setupMock func(repo *storage.MockInteractionRepository)
		wantErr   error
	}{
		{
			name: "actor at cap returns ErrMatchLimitReached without checking target",
			setupMock: func(repo *storage.MockInteractionRepository) {
				repo.EXPECT().CountActiveMatches(ctx, actorID, gomock.Nil()).Return(int64(2), nil)
			},
			wantErr: ErrMatchLimitReached,
		},
		{
			name: "actor at cap above limit returns ErrMatchLimitReached (grandfathered)",
			setupMock: func(repo *storage.MockInteractionRepository) {
				repo.EXPECT().CountActiveMatches(ctx, actorID, gomock.Nil()).Return(int64(5), nil)
			},
			wantErr: ErrMatchLimitReached,
		},
		{
			name: "actor under cap, target at cap returns ErrTargetMatchLimitReached",
			setupMock: func(repo *storage.MockInteractionRepository) {
				repo.EXPECT().CountActiveMatches(ctx, actorID, gomock.Nil()).Return(int64(1), nil)
				repo.EXPECT().CountActiveMatches(ctx, targetID, gomock.Nil()).Return(int64(2), nil)
			},
			wantErr: ErrTargetMatchLimitReached,
		},
		{
			name: "both at 1 active match returns nil (allowed to form match)",
			setupMock: func(repo *storage.MockInteractionRepository) {
				repo.EXPECT().CountActiveMatches(ctx, actorID, gomock.Nil()).Return(int64(1), nil)
				repo.EXPECT().CountActiveMatches(ctx, targetID, gomock.Nil()).Return(int64(1), nil)
			},
			wantErr: nil,
		},
		{
			name: "both at 0 active matches returns nil",
			setupMock: func(repo *storage.MockInteractionRepository) {
				repo.EXPECT().CountActiveMatches(ctx, actorID, gomock.Nil()).Return(int64(0), nil)
				repo.EXPECT().CountActiveMatches(ctx, targetID, gomock.Nil()).Return(int64(0), nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := storage.NewMockInteractionRepository(ctrl)
			tc.setupMock(repo)

			svc := newServiceWithRepo(t, repo)

			err := svc.enforceActiveMatchCap(ctx, nil, actorID, targetID)

			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Truef(t, errors.Is(err, tc.wantErr), "expected error %v, got %v", tc.wantErr, err)
		})
	}
}

func TestCreateSwipeNonMatchableVoiceNotePreservesVoiceMessageType(t *testing.T) {
	const (
		actorID        = "actor-1"
		targetID       = "target-1"
		clientMsgID    = "client-msg-1"
		promptID       = int64(42)
		voiceNoteURL   = "https://example.com/voice-note.m4a"
		mediaSeconds   = 8.5
		messageType    = constants.MessageTypeVoice
		expectedResult = ResultSent
	)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)
	tx := &fakeTx{}

	repo.EXPECT().CheckIfMatchable(ctx, actorID, targetID).Return(false, nil)
	repo.EXPECT().
		InsertSwipe(ctx, gomock.Any(), gomock.Nil()).
		DoAndReturn(func(_ context.Context, swipe entity.Swipe, _ *sql.Tx) error {
			require.False(t, swipe.Message.Valid)
			require.True(t, swipe.MessageType.Valid)
			assert.Equal(t, messageType, swipe.MessageType.String)
			require.True(t, swipe.VoicenoteURL.Valid)
			assert.Equal(t, voiceNoteURL, swipe.VoicenoteURL.String)
			require.True(t, swipe.IdempotencyKey.Valid)
			assert.Equal(t, clientMsgID, swipe.IdempotencyKey.String)

			gotSeconds, err := utils.NullDecimalToFloatPtr(swipe.MediaSeconds)
			require.NoError(t, err)
			require.NotNil(t, gotSeconds)
			assert.InDelta(t, mediaSeconds, *gotSeconds, 0.001)

			return nil
		})

	svc := &service{
		logger:          zaptest.NewLogger(t),
		uow:             &fakeUoW{tx: tx},
		interactionRepo: repo,
		discoverService: fakeDiscoverService{},
		hub:             fakeBroadcaster{},
	}

	result, err := svc.CreateSwipe(ctx, interactiondomain.Swipe{
		UserID:         actorID,
		TargetUserID:   targetID,
		Action:         constants.ActionLike,
		PromptID:       ptr(promptID),
		IdempotencyKey: ptr(clientMsgID),
		MessageType:    ptr(messageType),
		VoiceNoteURL:   ptr(voiceNoteURL),
		MediaSeconds:   ptr(mediaSeconds),
	})

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.True(t, tx.committed)
}

// TestGetLikesEmptyIncomingPopulatesViewerCounts covers the "no incoming
// likes" path: the per-id loop is skipped entirely so none of the per-like
// dependencies (safety, profile, discover) are touched, and the response
// still surfaces the viewer's active_matches_count + match_slot_limit.
//
// This also implicitly exercises CountActiveMatchesForUsers' empty-input
// early-return: it must be called with the empty (nil) slice and return an
// empty map without error.
func TestGetLikesEmptyIncomingPopulatesViewerCounts(t *testing.T) {
	const userID = "viewer-1"

	ctx := context.Background()

	cases := []struct {
		name              string
		viewerActiveCount int64
		wantSlotLimit     int64
	}{
		{name: "viewer at 0 matches", viewerActiveCount: 0, wantSlotLimit: 2},
		{name: "viewer at 1 match", viewerActiveCount: 1, wantSlotLimit: 2},
		{name: "viewer at 2 matches (at cap)", viewerActiveCount: 2, wantSlotLimit: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := storage.NewMockInteractionRepository(ctrl)

			repo.EXPECT().GetIncomingLikes(ctx, userID, 5, 0).Return(nil, nil)
			repo.EXPECT().CountActiveMatchesForUsers(ctx, gomock.Nil()).Return(map[string]int64{}, nil)
			repo.EXPECT().GetWatchedUserIDs(ctx, userID).Return(map[string]struct{}{}, nil)
			repo.EXPECT().CountActiveMatches(ctx, userID, gomock.Nil()).Return(tc.viewerActiveCount, nil)

			svc := newServiceWithRepo(t, repo)

			likes, err := svc.GetLikes(ctx, userID, "incoming", 0, 5)

			require.NoError(t, err)
			assert.Equal(t, tc.viewerActiveCount, likes.ActiveMatchesCount)
			assert.Equal(t, tc.wantSlotLimit, likes.MatchSlotLimit)
			assert.Empty(t, likes.FreeToMatch)
			assert.Empty(t, likes.SlotsFull)
		})
	}
}

// TestGetLikesInvalidDirectionShortCircuits guards the early-return path so
// no repo calls are made on bad input (mirrors the production behaviour and
// catches accidental reordering of the direction switch).
func TestGetLikesInvalidDirectionShortCircuits(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)

	svc := newServiceWithRepo(t, repo)

	likes, err := svc.GetLikes(context.Background(), "viewer-1", "outgoing", 0, 5)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDirection)
	assert.Empty(t, likes.FreeToMatch)
	assert.Empty(t, likes.SlotsFull)
}

func TestCreateSwipeMarksDailyPickDecided(t *testing.T) {
	const (
		actorID  = "actor-1"
		targetID = "target-1"
	)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)
	dailyPicks := dailypicks.NewMockService(ctrl)
	tx := &fakeTx{}

	repo.EXPECT().CheckIfMatchable(ctx, actorID, targetID).Return(false, nil)
	repo.EXPECT().InsertSwipe(ctx, gomock.Any(), gomock.Nil()).Return(nil)
	dailyPicks.EXPECT().
		MarkDecided(ctx, gomock.Nil(), actorID, targetID, constants.ActionPass).
		Return(nil)

	svc := &service{
		logger:            zaptest.NewLogger(t),
		uow:               &fakeUoW{tx: tx},
		interactionRepo:   repo,
		discoverService:   fakeDiscoverService{},
		hub:               fakeBroadcaster{},
		dailyPicksService: dailyPicks,
	}

	result, err := svc.CreateSwipe(ctx, interactiondomain.Swipe{
		UserID:       actorID,
		TargetUserID: targetID,
		Action:       constants.ActionPass,
	})

	require.NoError(t, err)
	assert.Equal(t, ResultPassed, result)
	assert.True(t, tx.committed)
}

func TestCreateSwipeMarkDecidedErrorDoesNotFailSwipe(t *testing.T) {
	const (
		actorID  = "actor-1"
		targetID = "target-1"
	)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)
	dailyPicks := dailypicks.NewMockService(ctrl)
	tx := &fakeTx{}

	repo.EXPECT().CheckIfMatchable(ctx, actorID, targetID).Return(false, nil)
	repo.EXPECT().InsertSwipe(ctx, gomock.Any(), gomock.Nil()).Return(nil)
	dailyPicks.EXPECT().
		MarkDecided(ctx, gomock.Nil(), actorID, targetID, constants.ActionPass).
		Return(errors.New("daily picks unavailable"))

	svc := &service{
		logger:            zaptest.NewLogger(t),
		uow:               &fakeUoW{tx: tx},
		interactionRepo:   repo,
		discoverService:   fakeDiscoverService{},
		hub:               fakeBroadcaster{},
		dailyPicksService: dailyPicks,
	}

	result, err := svc.CreateSwipe(ctx, interactiondomain.Swipe{
		UserID:       actorID,
		TargetUserID: targetID,
		Action:       constants.ActionPass,
	})

	require.NoError(t, err)
	assert.Equal(t, ResultPassed, result)
	assert.True(t, tx.committed)
}

// TestMaxActiveMatchesIsTwo pins the spec contract: HAE-411 requires a hard
// cap of 2. If anyone bumps this, they should also revisit the FE copy and
// the messages exposed by the swipes handler.
func TestMaxActiveMatchesIsTwo(t *testing.T) {
	assert.Equal(t, int64(2), constants.MaxActiveMatches, "match cap must remain 2 per HAE-411")
}

func ptr[T any](value T) *T {
	return &value
}
