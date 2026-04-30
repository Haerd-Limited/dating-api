package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile"
)

func newOnboardingServiceForTest(t *testing.T, lookupRepo lookupstorage.LookupRepository) *onboardingService {
	t.Helper()

	return &onboardingService{
		logger:     zaptest.NewLogger(t),
		lookupRepo: lookupRepo,
	}
}

// buildVoicePrompts builds a slice of valid VoicePrompt entries with the given
// prompt_type IDs, using sequential positions starting at 1.
func buildVoicePrompts(promptTypeIDs ...int16) []domain.VoicePrompt {
	out := make([]domain.VoicePrompt, 0, len(promptTypeIDs))
	for i, id := range promptTypeIDs {
		out = append(out, domain.VoicePrompt{
			URL:          "https://example.com/audio.m4a",
			PromptType:   id,
			Position:     int16(i + 1),
			WaveformData: []float32{0.1, 0.2, 0.3},
		})
	}

	return out
}

func TestValidatePrompts(t *testing.T) {
	ctx := context.Background()
	coreIDs := []int16{1, 2, 3, 4, 5}

	type setup func(repo *lookupstorage.MockLookupRepository)

	cases := []struct {
		name      string
		prompts   []domain.VoicePrompt
		setupMock setup
		wantErr   error
	}{
		{
			name:    "missing one core prompt returns ErrMissingRequiredCorePrompts",
			prompts: buildVoicePrompts(1, 2, 3, 4, 99),
			setupMock: func(repo *lookupstorage.MockLookupRepository) {
				repo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: profile.ErrMissingRequiredCorePrompts,
		},
		{
			name:    "all 5 core + 0 optional succeeds",
			prompts: buildVoicePrompts(1, 2, 3, 4, 5),
			setupMock: func(repo *lookupstorage.MockLookupRepository) {
				repo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: nil,
		},
		{
			name:    "all 5 core + 5 optional (10 total) succeeds",
			prompts: buildVoicePrompts(1, 2, 3, 4, 5, 100, 101, 102, 103, 104),
			setupMock: func(repo *lookupstorage.MockLookupRepository) {
				repo.EXPECT().GetCorePromptTypeIDs(ctx).Return(coreIDs, nil)
			},
			wantErr: nil,
		},
		{
			name:    "11 prompts returns ErrTooManyPromptsProvided before core check",
			prompts: buildVoicePrompts(1, 2, 3, 4, 5, 100, 101, 102, 103, 104, 105),
			// no mock expectation: count check should reject before lookupRepo is hit
			setupMock: func(repo *lookupstorage.MockLookupRepository) {},
			wantErr:   ErrTooManyPromptsProvided,
		},
		{
			name:      "4 prompts returns ErrNotEnoughPromptsProvided before core check",
			prompts:   buildVoicePrompts(1, 2, 3, 4),
			setupMock: func(repo *lookupstorage.MockLookupRepository) {},
			wantErr:   ErrNotEnoughPromptsProvided,
		},
		{
			name:      "empty prompts returns ErrMissingPrompts before core check",
			prompts:   nil,
			setupMock: func(repo *lookupstorage.MockLookupRepository) {},
			wantErr:   ErrMissingPrompts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := lookupstorage.NewMockLookupRepository(ctrl)
			tc.setupMock(repo)

			svc := newOnboardingServiceForTest(t, repo)

			err := svc.validatePrompts(ctx, domain.Prompts{
				UserID:          "user-123",
				UploadedPrompts: tc.prompts,
			})

			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Truef(t, errors.Is(err, tc.wantErr), "expected error %v, got %v", tc.wantErr, err)
		})
	}
}

func TestValidatePromptsPropagatesLookupRepoError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")

	ctrl := gomock.NewController(t)
	repo := lookupstorage.NewMockLookupRepository(ctrl)
	repo.EXPECT().GetCorePromptTypeIDs(ctx).Return(nil, repoErr)

	svc := newOnboardingServiceForTest(t, repo)

	err := svc.validatePrompts(ctx, domain.Prompts{
		UserID:          "user-123",
		UploadedPrompts: buildVoicePrompts(1, 2, 3, 4, 5),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}
