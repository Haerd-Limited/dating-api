package consent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	consentstorage "github.com/Haerd-Limited/dating-api/internal/consent/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

func TestGetMissingMandatory(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := consentstorage.NewMockRepository(ctrl)
	svc := NewService(zaptest.NewLogger(t), repo).(*service)

	ctx := context.Background()
	userID := "user-1"

	repo.EXPECT().GetMissingMandatory(ctx, userID, constants.MandatoryConsentTypes, gomock.Any()).
		Return([]string{constants.ConsentTypePrivacyPolicy, constants.ConsentTypeTermsOfService}, nil)

	missing, err := svc.GetMissingMandatory(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, []string{constants.ConsentTypePrivacyPolicy, constants.ConsentTypeTermsOfService}, missing)

	// Cached result should not hit repo again.
	missing, err = svc.GetMissingMandatory(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, []string{constants.ConsentTypePrivacyPolicy, constants.ConsentTypeTermsOfService}, missing)
}

func TestRecordInvalidatesCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := consentstorage.NewMockRepository(ctrl)
	svc := NewService(zaptest.NewLogger(t), repo).(*service)

	ctx := context.Background()
	userID := "user-1"

	svc.cache.Store(userID, cacheEntry{
		missing:   []string{constants.ConsentTypePrivacyPolicy},
		expiresAt: time.Now().Add(time.Minute),
	})

	repo.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
	repo.EXPECT().GetMissingMandatory(ctx, userID, constants.MandatoryConsentTypes, gomock.Any()).
		Return([]string{}, nil)

	err := svc.Record(ctx, domain.RecordRequest{
		UserID:   userID,
		Type:     constants.ConsentTypePrivacyPolicy,
		Version:  constants.CurrentPrivacyPolicyVersion,
		Accepted: true,
	})
	require.NoError(t, err)

	missing, err := svc.GetMissingMandatory(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, missing)
}

func TestRevokeAddsMissingType(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := consentstorage.NewMockRepository(ctrl)
	svc := NewService(zaptest.NewLogger(t), repo).(*service)

	ctx := context.Background()
	userID := "user-1"

	repo.EXPECT().Revoke(ctx, userID, constants.ConsentTypeTermsOfService).Return(nil)
	repo.EXPECT().GetMissingMandatory(ctx, userID, constants.MandatoryConsentTypes, gomock.Any()).
		Return([]string{constants.ConsentTypeTermsOfService}, nil)

	err := svc.Revoke(ctx, userID, constants.ConsentTypeTermsOfService)
	require.NoError(t, err)

	missing, err := svc.GetMissingMandatory(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, []string{constants.ConsentTypeTermsOfService}, missing)
}

func TestRecordInvalidConsentType(t *testing.T) {
	svc := NewService(zaptest.NewLogger(t), nil)

	err := svc.Record(context.Background(), domain.RecordRequest{
		UserID:  "user-1",
		Type:    "marketing",
		Version: "2026-04-30",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidConsentType))
}
