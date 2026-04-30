package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile/storage"
)

func buildPromptUpdates(promptTypeIDs ...int16) []domain.VoicePromptUpdate {
	out := make([]domain.VoicePromptUpdate, 0, len(promptTypeIDs))
	for i, id := range promptTypeIDs {
		out = append(out, domain.VoicePromptUpdate{
			URL:          "https://example.com/audio.m4a",
			PromptTypeID: id,
			Position:     int16(i + 1),
			WaveformData: []float32{0.1, 0.2, 0.3},
		})
	}

	return out
}

func newProfileServiceForTest(
	t *testing.T,
	profileRepo storage.ProfileRepository,
	lookupRepo lookupstorage.LookupRepository,
) *service {
	t.Helper()

	return &service{
		logger:      zaptest.NewLogger(t),
		profileRepo: profileRepo,
		lookupRepo:  lookupRepo,
	}
}

func TestUpsertUserPromptsCorePromptValidation(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	coreIDs := []int16{1, 2, 3, 4, 5}

	type setup func(profileRepo *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository)

	cases := []struct {
		name      string
		prompts   []domain.VoicePromptUpdate
		setupMock setup
		wantErr   error
	}{
		{
			name:    "missing one core prompt returns ErrMissingRequiredCorePrompts",
			prompts: buildPromptUpdates(1, 2, 3, 4, 99),
			setupMock: func(_ *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository) {
				lookupRepo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: ErrMissingRequiredCorePrompts,
		},
		{
			name:    "all 5 core + 0 optional succeeds",
			prompts: buildPromptUpdates(1, 2, 3, 4, 5),
			setupMock: func(profileRepo *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository) {
				lookupRepo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
				profileRepo.EXPECT().UpsertUserPrompts(ctx, userID, gomock.Any(), gomock.Nil()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "all 5 core + 5 optional (10 total) succeeds",
			prompts: buildPromptUpdates(1, 2, 3, 4, 5, 100, 101, 102, 103, 104),
			setupMock: func(profileRepo *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository) {
				lookupRepo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
				profileRepo.EXPECT().UpsertUserPrompts(ctx, userID, gomock.Any(), gomock.Nil()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "11 prompts returns ErrTooManyPromptsProvided",
			prompts: buildPromptUpdates(1, 2, 3, 4, 5, 100, 101, 102, 103, 104, 105),
			setupMock: func(_ *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository) {
				lookupRepo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: ErrTooManyPromptsProvided,
		},
		{
			name:    "empty prompts returns ErrMissingPrompts",
			prompts: nil,
			setupMock: func(_ *storage.MockProfileRepository, lookupRepo *lookupstorage.MockLookupRepository) {
				lookupRepo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: ErrMissingPrompts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			profileRepo := storage.NewMockProfileRepository(ctrl)
			lookupRepo := lookupstorage.NewMockLookupRepository(ctrl)
			tc.setupMock(profileRepo, lookupRepo)

			svc := newProfileServiceForTest(t, profileRepo, lookupRepo)

			err := svc.UpsertUserPrompts(ctx, userID, tc.prompts)

			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Truef(t, errors.Is(err, tc.wantErr), "expected error %v, got %v", tc.wantErr, err)
		})
	}
}

// TestValidateUserPromptsUpsertCoreCheckOnly is the unit-level proof of the
// validator's contract. The grandfather guarantee comes from the fact that
// callers (service.UpdateProfile and service.UpsertUserPrompts) only invoke
// this validator when a non-empty voice_prompts payload is being written —
// see the `if len(up.VoicePrompts) > 0` gate in UpdateProfile. A basics-only
// edit (VoicePrompts == nil) never reaches this validator, so users with
// incomplete stored prompts can still update non-prompt fields.
func TestValidateUserPromptsUpsertCoreCheckOnly(t *testing.T) {
	coreIDs := []int16{1, 2, 3, 4, 5}

	t.Run("all core present passes", func(t *testing.T) {
		err := validateUserPromptsUpsert(buildPromptUpdates(1, 2, 3, 4, 5), coreIDs)
		require.NoError(t, err)
	})

	t.Run("missing core prompt 5 returns ErrMissingRequiredCorePrompts", func(t *testing.T) {
		err := validateUserPromptsUpsert(buildPromptUpdates(1, 2, 3, 4, 99), coreIDs)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingRequiredCorePrompts)
	})

	t.Run("nil coreIDs (no core prompts configured) passes any non-empty set", func(t *testing.T) {
		err := validateUserPromptsUpsert(buildPromptUpdates(7, 8, 9, 10, 11), nil)
		require.NoError(t, err)
	})
}
