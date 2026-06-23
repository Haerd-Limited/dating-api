package consent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/api/consent/dto"
	internalconsent "github.com/Haerd-Limited/dating-api/internal/consent"
	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

func TestRecordConsentHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := internalconsent.NewMockService(ctrl)
	h := NewConsentHandler(zaptest.NewLogger(t), svc)

	body, err := json.Marshal(dto.RecordConsentRequest{
		Type:     constants.ConsentTypePrivacyPolicy,
		Version:  constants.CurrentPrivacyPolicyVersion,
		Accepted: true,
	})
	require.NoError(t, err)

	svc.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req domain.RecordRequest) error {
			assert.Equal(t, constants.ConsentTypePrivacyPolicy, req.Type)
			assert.Equal(t, "user-1", req.UserID)

			return nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consents", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1"))
	rec := httptest.NewRecorder()
	h.Record().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecordConsentInvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := internalconsent.NewMockService(ctrl)
	h := NewConsentHandler(zaptest.NewLogger(t), svc)

	body, err := json.Marshal(dto.RecordConsentRequest{
		Type:     "marketing",
		Version:  constants.CurrentPrivacyPolicyVersion,
		Accepted: true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consents", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1"))
	rec := httptest.NewRecorder()
	h.Record().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRecordConsentUnauthorized(t *testing.T) {
	h := NewConsentHandler(zaptest.NewLogger(t), nil)

	body, err := json.Marshal(dto.RecordConsentRequest{
		Type:     constants.ConsentTypePrivacyPolicy,
		Version:  constants.CurrentPrivacyPolicyVersion,
		Accepted: true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consents", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Record().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRecordConsentServiceInvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := internalconsent.NewMockService(ctrl)
	h := NewConsentHandler(zaptest.NewLogger(t), svc)

	body, err := json.Marshal(dto.RecordConsentRequest{
		Type:     constants.ConsentTypePrivacyPolicy,
		Version:  constants.CurrentPrivacyPolicyVersion,
		Accepted: true,
	})
	require.NoError(t, err)

	svc.EXPECT().Record(gomock.Any(), gomock.Any()).Return(internalconsent.ErrInvalidConsentType)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consents", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1"))
	rec := httptest.NewRecorder()
	h.Record().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
