package interaction

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
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

// TestMaxActiveMatchesIsTwo pins the spec contract: HAE-411 requires a hard
// cap of 2. If anyone bumps this, they should also revisit the FE copy and
// the messages exposed by the swipes handler.
func TestMaxActiveMatchesIsTwo(t *testing.T) {
	assert.Equal(t, int64(2), constants.MaxActiveMatches, "match cap must remain 2 per HAE-411")
}
