package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/auth/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
)

// update tests to follow Andres advice

func TestRefreshHandler(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLog := zaptest.NewLogger(t)

	testCases := []struct {
		name  string
		ctx   func(ctx context.Context) context.Context
		input dto.RefreshRequest
		mock  func(
			mockService *auth.MockService,
		)
		wantStatus int
		wantBody   any
	}{
		{
			name:       "successful refresh",
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"message":       "Tokens refreshed successfully",
				"access_token":  "newAccessToken",
				"refresh_token": "newRefreshToken",
			},
			input: dto.RefreshRequest{
				RefreshToken: "testRefreshToken",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockService,
			) {
				mockService.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(
					&domain.AuthResult{
						RefreshToken: "newRefreshToken",
						AccessToken:  "newAccessToken",
					},
					nil,
				)
			},
		},
		{
			name:       "refresh token missing",
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": InvalidRefreshTokenMsg,
			},
			input: dto.RefreshRequest{},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockService,
			) {
			},
		},
		{
			name:       "a service error",
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]interface{}{
				"message": messages.InternalServerErrorMsg,
			},
			input: dto.RefreshRequest{
				RefreshToken: "testRefreshToken",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockService,
			) {
				mockService.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(
					nil,
					assert.AnError,
				)
			},
		},
		{
			name:       "token revoked or missing",
			wantStatus: http.StatusUnauthorized,
			wantBody: map[string]interface{}{
				"message": TokenRevokedOrExpiredMsg,
			},
			input: dto.RefreshRequest{
				RefreshToken: "testRefreshToken",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockService,
			) {
				mockService.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(
					nil,
					auth.ErrRefreshTokenRevoked,
				)
			},
		},
	}
	// Run your test cases
	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				mockService := auth.NewMockService(ctrl)

				tc.mock(mockService)

				body, err := json.Marshal(tc.input)
				require.NoError(t, err)

				rctx := chi.NewRouteContext()

				req, err := http.NewRequestWithContext(
					context.WithValue(tc.ctx(context.Background()), chi.RouteCtxKey, rctx),
					http.MethodPost,
					"/api/v1/auth/refresh",
					bytes.NewReader(body),
				)
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				h := NewAuthHandler(mockLog, mockService)
				h.Refresh().ServeHTTP(recorder, req)

				expected, err := json.Marshal(tc.wantBody)
				require.NoError(t, err)

				assert.Equal(t, tc.wantStatus, recorder.Code)

				var actual map[string]interface{}
				err = json.Unmarshal(recorder.Body.Bytes(), &actual)
				require.NoError(t, err)

				var expectedJSON map[string]interface{}
				err = json.Unmarshal(expected, &expectedJSON)
				require.NoError(t, err)

				assert.Equal(t, expectedJSON, actual)

				satisfied := ctrl.Satisfied()
				require.True(t, satisfied, "mock expectations were not satisfied")
			},
		)
	}
}

func TestLogoutHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLog := zaptest.NewLogger(t)

	testCases := []struct {
		name       string
		ctx        func(ctx context.Context) context.Context
		input      dto.LogoutRequest
		mock       func(mockService *auth.MockService)
		wantStatus int
		wantBody   map[string]interface{}
	}{
		{
			name: "successful logout",
			input: dto.LogoutRequest{
				RefreshToken: "validRefreshToken",
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"message": "Logged out successfully",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(mockService *auth.MockService) {
				mockService.EXPECT().
					RevokeRefreshToken(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:       "missing refresh token",
			input:      dto.LogoutRequest{},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": InvalidRefreshTokenMsg,
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(mockService *auth.MockService) {
				// no call expected
			},
		},
		{
			name: "internal error during logout",
			input: dto.LogoutRequest{
				RefreshToken: "token",
			},
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]interface{}{
				"message": messages.InternalServerErrorMsg,
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(mockService *auth.MockService) {
				mockService.EXPECT().
					RevokeRefreshToken(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
		},
		{
			name: "token already revoked",
			input: dto.LogoutRequest{
				RefreshToken: "revokedToken",
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"message": TokenRevokedOrExpiredMsg,
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(mockService *auth.MockService) {
				mockService.EXPECT().
					RevokeRefreshToken(gomock.Any(), gomock.Any()).
					Return(auth.ErrRefreshTokenAlreadyRevoked)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := auth.NewMockService(ctrl)
			tc.mock(mockService)

			body, err := json.Marshal(tc.input)
			require.NoError(t, err)

			rctx := chi.NewRouteContext()
			req, err := http.NewRequestWithContext(
				context.WithValue(tc.ctx(context.Background()), chi.RouteCtxKey, rctx),
				http.MethodPost,
				"/api/v1/auth/logout",
				bytes.NewReader(body),
			)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			h := NewAuthHandler(mockLog, mockService)
			h.Logout().ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantStatus, recorder.Code)

			var actual map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.wantBody, actual)

			satisfied := ctrl.Satisfied()
			require.True(t, satisfied, "mock expectations were not satisfied")
		})
	}
}
