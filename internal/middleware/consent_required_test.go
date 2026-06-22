package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	internalconsent "github.com/Haerd-Limited/dating-api/internal/consent"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

func TestConsentRequired(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		enabled    bool
		setupMock  func(*internalconsent.MockService)
		withUser   bool
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:       "disabled passes through",
			enabled:    false,
			withUser:   true,
			setupMock:  func(_ *internalconsent.MockService) {},
			wantStatus: http.StatusOK,
		},
		{
			name:     "enabled missing both consents",
			enabled:  true,
			withUser: true,
			setupMock: func(svc *internalconsent.MockService) {
				svc.EXPECT().GetMissingMandatory(gomock.Any(), "user-1").
					Return([]string{constants.ConsentTypePrivacyPolicy, constants.ConsentTypeTermsOfService}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody: map[string]any{
				"error":   "consent_required",
				"missing": []any{constants.ConsentTypePrivacyPolicy, constants.ConsentTypeTermsOfService},
			},
		},
		{
			name:     "enabled fully consented",
			enabled:  true,
			withUser: true,
			setupMock: func(svc *internalconsent.MockService) {
				svc.EXPECT().GetMissingMandatory(gomock.Any(), "user-1").Return([]string{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "enabled one revoked",
			enabled:  true,
			withUser: true,
			setupMock: func(svc *internalconsent.MockService) {
				svc.EXPECT().GetMissingMandatory(gomock.Any(), "user-1").
					Return([]string{constants.ConsentTypeTermsOfService}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody: map[string]any{
				"error":   "consent_required",
				"missing": []any{constants.ConsentTypeTermsOfService},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := internalconsent.NewMockService(ctrl)
			tc.setupMock(svc)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				w.WriteHeader(http.StatusOK)
			})

			handler := ConsentRequired(svc, tc.enabled, zaptest.NewLogger(t))(next)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/intro", nil)

			ctx := req.Context()
			if tc.withUser {
				ctx = context.WithValue(ctx, commoncontext.UserIDKey, "user-1")
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req.WithContext(ctx))

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				assert.True(t, nextCalled)
				return
			}

			assert.False(t, nextCalled)

			if tc.wantBody != nil {
				var body map[string]any

				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tc.wantBody["error"], body["error"])
				assert.Equal(t, tc.wantBody["missing"], body["missing"])
			}
		})
	}
}
