package compatibility

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/storage"
)

func TestCalculateCompatibilityPercent(t *testing.T) {
	tests := []struct {
		name     string
		earnedAB int
		totalAB  int
		earnedBA int
		totalBA  int
		want     int
	}{
		{
			name:     "perfect match",
			earnedAB: 10,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  20,
			want:     100,
		},
		{
			name:     "no match",
			earnedAB: 0,
			totalAB:  10,
			earnedBA: 0,
			totalBA:  20,
			want:     0,
		},
		{
			name:     "one sided half match",
			earnedAB: 5,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  20,
			want:     71,
		},
		{
			name:     "zero totals default to full satisfaction",
			earnedAB: 0,
			totalAB:  0,
			earnedBA: 0,
			totalBA:  0,
			want:     100,
		},
		{
			name:     "clamps above one hundred",
			earnedAB: 20,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  10,
			want:     100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCompatibilityPercent(tt.earnedAB, tt.totalAB, tt.earnedBA, tt.totalBA)
			if got != tt.want {
				t.Fatalf("calculateCompatibilityPercent() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsQuestionPacksComplete(t *testing.T) {
	const userID = "user-compat-test"

	cases := []struct {
		name        string
		setupMock   func(repo *storage.MockCompatibilityRepository)
		want        bool
		wantErr     bool
		errContains string
	}{
		{
			name: "40 answered 40 active returns true",
			setupMock: func(repo *storage.MockCompatibilityRepository) {
				repo.EXPECT().CountQuestions(gomock.Any(), gomock.Nil()).Return(40, nil)
				repo.EXPECT().CountUserAnswers(gomock.Any(), userID).Return(40, nil)
			},
			want: true,
		},
		{
			name: "39 answered 40 active returns false",
			setupMock: func(repo *storage.MockCompatibilityRepository) {
				repo.EXPECT().CountQuestions(gomock.Any(), gomock.Nil()).Return(40, nil)
				repo.EXPECT().CountUserAnswers(gomock.Any(), userID).Return(39, nil)
			},
			want: false,
		},
		{
			name: "41 answered 40 active returns true when question deactivated",
			setupMock: func(repo *storage.MockCompatibilityRepository) {
				repo.EXPECT().CountQuestions(gomock.Any(), gomock.Nil()).Return(40, nil)
				repo.EXPECT().CountUserAnswers(gomock.Any(), userID).Return(41, nil)
			},
			want: true,
		},
		{
			name: "CountQuestions error propagated",
			setupMock: func(repo *storage.MockCompatibilityRepository) {
				repo.EXPECT().CountQuestions(gomock.Any(), gomock.Nil()).Return(0, errors.New("db down"))
			},
			wantErr:     true,
			errContains: "count active questions",
		},
		{
			name: "CountUserAnswers error propagated",
			setupMock: func(repo *storage.MockCompatibilityRepository) {
				repo.EXPECT().CountQuestions(gomock.Any(), gomock.Nil()).Return(40, nil)
				repo.EXPECT().CountUserAnswers(gomock.Any(), userID).Return(0, errors.New("db down"))
			},
			wantErr:     true,
			errContains: "count user answers",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := storage.NewMockCompatibilityRepository(ctrl)
			tc.setupMock(repo)

			svc := NewCompatibilityService(zaptest.NewLogger(t), repo)
			got, err := svc.IsQuestionPacksComplete(context.Background(), userID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errContains))

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
