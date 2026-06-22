package preference

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/preference/storage"
)

func TestIsAnalyticsOptedOut(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := storage.NewMockPreferenceRepository(ctrl)
	svc := NewPreferenceService(zaptest.NewLogger(t), repo)

	ctx := context.Background()

	repo.EXPECT().IsAnalyticsOptedOut(ctx, "user-1").Return(false, nil)
	optedOut, err := svc.IsAnalyticsOptedOut(ctx, "user-1")
	require.NoError(t, err)
	assert.False(t, optedOut)

	repo.EXPECT().IsAnalyticsOptedOut(ctx, "user-2").Return(true, nil)
	optedOut, err = svc.IsAnalyticsOptedOut(ctx, "user-2")
	require.NoError(t, err)
	assert.True(t, optedOut)
}

func TestSetAnalyticsOptOut(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := storage.NewMockPreferenceRepository(ctrl)
	svc := NewPreferenceService(zaptest.NewLogger(t), repo)

	ctx := context.Background()

	repo.EXPECT().SetAnalyticsOptOut(ctx, "user-1", true).Return(nil)
	err := svc.SetAnalyticsOptOut(ctx, "user-1", true)
	require.NoError(t, err)

	repo.EXPECT().SetAnalyticsOptOut(ctx, "user-1", false).Return(nil)
	err = svc.SetAnalyticsOptOut(ctx, "user-1", false)
	require.NoError(t, err)
}

func TestIsAnalyticsOptedOutRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := storage.NewMockPreferenceRepository(ctrl)
	svc := NewPreferenceService(zaptest.NewLogger(t), repo)

	repo.EXPECT().IsAnalyticsOptedOut(gomock.Any(), "user-1").Return(false, errors.New("db down"))
	_, err := svc.IsAnalyticsOptedOut(context.Background(), "user-1")
	require.Error(t, err)
}
