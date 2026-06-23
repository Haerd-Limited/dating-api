package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth"
)

func newRequestWithBody(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rctx := chi.NewRouteContext()
	req, err := http.NewRequestWithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, rctx),
		method,
		target,
		bytes.NewReader(raw),
	)
	require.NoError(t, err)

	return req
}

// TestVerifyCodeHandlerAccountGate proves the banned/suspended errors are
// surfaced as 403 instead of being swallowed into the anti-enumeration 200.
func TestVerifyCodeHandlerAccountGate(t *testing.T) {
	until := time.Now().UTC().Add(48 * time.Hour)

	cases := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:       "banned returns 403",
			serviceErr: auth.ErrAccountBanned,
			wantStatus: http.StatusForbidden,
			wantBody:   map[string]any{"error": "account_banned"},
		},
		{
			name:       "suspended returns 403 with until",
			serviceErr: &auth.SuspendedAccountError{Until: until},
			wantStatus: http.StatusForbidden,
			wantBody: map[string]any{
				"error": "account_suspended",
				"until": until.Format(time.RFC3339),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := auth.NewMockService(ctrl)

			mockService.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).Return(nil, tc.serviceErr)

			req := newRequestWithBody(t, http.MethodPost, "/api/v1/auth/verify", dto.VerifyCodeRequest{
				Channel: "sms",
				Purpose: "login",
				Code:    "123456",
			})

			recorder := httptest.NewRecorder()
			h := NewAuthHandler(zaptest.NewLogger(t), mockService)
			h.VerifyCode().ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantStatus, recorder.Code)

			var actual map[string]any

			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actual))
			assert.Equal(t, tc.wantBody, actual)
		})
	}
}

// TestRefreshHandlerAccountGate proves a banned/suspended user is rejected with
// 403 and that no tokens are minted in the response body.
func TestRefreshHandlerAccountGate(t *testing.T) {
	until := time.Now().UTC().Add(48 * time.Hour)

	cases := []struct {
		name       string
		serviceErr error
		wantBody   map[string]any
	}{
		{
			name:       "banned returns 403",
			serviceErr: auth.ErrAccountBanned,
			wantBody:   map[string]any{"error": "account_banned"},
		},
		{
			name:       "suspended returns 403 with until",
			serviceErr: &auth.SuspendedAccountError{Until: until},
			wantBody: map[string]any{
				"error": "account_suspended",
				"until": until.Format(time.RFC3339),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := auth.NewMockService(ctrl)

			mockService.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(nil, tc.serviceErr)

			req := newRequestWithBody(t, http.MethodPost, "/api/v1/auth/refresh", dto.RefreshRequest{
				RefreshToken: "some-refresh-token",
			})

			recorder := httptest.NewRecorder()
			h := NewAuthHandler(zaptest.NewLogger(t), mockService)
			h.Refresh().ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusForbidden, recorder.Code)

			var actual map[string]any

			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actual))
			assert.Equal(t, tc.wantBody, actual)

			_, hasAccess := actual["access_token"]
			_, hasRefresh := actual["refresh_token"]

			assert.False(t, hasAccess, "no access token should be minted")
			assert.False(t, hasRefresh, "no refresh token should be minted")
		})
	}
}
