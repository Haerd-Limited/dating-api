package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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
	"github.com/Haerd-Limited/dating-api/internal/user"
	userDomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
)

// todo: update tests to follow Andres advice

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name       string
		setupCtx   func() context.Context
		buildBody  func() (io.Reader, string)
		mock       func(mockService *auth.MockAuthService)
		wantStatus int
		wantBody   map[string]interface{}
	}{
		{
			name: "successful registration",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("username", "lionellegendz")
				_ = w.WriteField("email", "lionel@gmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970-01-01")
				_ = w.WriteField("gender", "male")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock: func(mockService *auth.MockAuthService) {
				mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(
					&domain.AuthTokensAndUser{
						AccessToken:  "testAccessToken",
						RefreshToken: "testRefreshToken",
						User: userDomain.User{
							FullName: "lionel wilson",
							Username: "lionellegendz",
							Email:    "lionel@gmail.com",
						},
					}, nil,
				)
			},
			wantStatus: http.StatusCreated,
			wantBody: map[string]interface{}{
				"message":       "Registration successful",
				"access_token":  "testAccessToken",
				"refresh_token": "testRefreshToken",
				"user_details": map[string]interface{}{
					"username":  "lionellegendz",
					"email":     "lionel@gmail.com",
					"full_name": "lionel wilson",
				},
			},
		},
		{
			name: "missing username",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("email", "lionel@gmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970-01-01")
				_ = w.WriteField("gender", "male")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock:       func(mockService *auth.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": MissingRequiredFieldMsg,
			},
		},
		{
			name: "invalid email",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("username", "lionellegendz")
				_ = w.WriteField("email", "lionelgmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970-01-01")
				_ = w.WriteField("gender", "male")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock: func(mockService *auth.MockAuthService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": InvalidEmailMsg,
			},
		},
		{
			name: "invalid dob",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("username", "lionellegendz")
				_ = w.WriteField("email", "lionel@gmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970/01/01")
				_ = w.WriteField("gender", "male")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock: func(mockService *auth.MockAuthService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": messages.InvalidDobMsg,
			},
		},
		{
			name: "invalid gender",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("username", "lionellegendz")
				_ = w.WriteField("email", "lionel@gmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970-01-01")
				_ = w.WriteField("gender", "other")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock: func(mockService *auth.MockAuthService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": messages.InvalidGenderMsg,
			},
		},
		{
			name: "service error",
			setupCtx: func() context.Context {
				return context.Background()
			},
			buildBody: func() (io.Reader, string) {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("full_name", "lionel wilson")
				_ = w.WriteField("username", "lionellegendz")
				_ = w.WriteField("email", "lionel@gmail.com")
				_ = w.WriteField("password", "testPassword")
				_ = w.WriteField("date_of_birth", "1970-01-01")
				_ = w.WriteField("gender", "male")
				_ = w.WriteField("bio", "Ceo of @haerd")
				_ = w.Close()
				return &buf, w.FormDataContentType()
			},
			mock: func(mockService *auth.MockAuthService) {
				mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(
					nil, assert.AnError,
				)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]interface{}{
				"message": messages.InternalServerErrorMsg,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := auth.NewMockAuthService(ctrl)
			tc.mock(mockService)

			body, contentType := tc.buildBody()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
			req = req.WithContext(tc.setupCtx())
			req.Header.Set("Content-Type", contentType)

			recorder := httptest.NewRecorder()
			authhandler := NewAuthHandler(logger, mockService)
			authhandler.Register().ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantStatus, recorder.Code)

			var actual map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.wantBody, actual)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLog := zaptest.NewLogger(t)

	testCases := []struct {
		name  string
		ctx   func(ctx context.Context) context.Context
		input dto.LoginRequest
		mock  func(
			mockService *auth.MockAuthService,
		)
		wantStatus int
		wantBody   any
	}{
		{
			name:       "successful login",
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"message":       "Login successful",
				"access_token":  "testAccessToken",
				"refresh_token": "testRefreshToken",
				"user_details": map[string]interface{}{
					"username":  "lionellegendz",
					"email":     "lionel@gmail.com",
					"full_name": "lionel wilson",
				},
			},
			input: dto.LoginRequest{
				Email:    "lionel@gmail.com",
				Password: "testPassword",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockAuthService,
			) {
				mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(
					&domain.AuthTokensAndUser{
						RefreshToken: "testRefreshToken",
						AccessToken:  "testAccessToken",
						User: userDomain.User{
							FullName: "lionel wilson",
							Username: "lionellegendz",
							Email:    "lionel@gmail.com",
						},
					},
					nil,
				)
			},
		},
		{
			name:       "email or password missing",
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": InvalidLoginInputMsg,
			},
			input: dto.LoginRequest{
				Password: "testPassword",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockAuthService,
			) {
			},
		},
		{
			name:       "a service error",
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]interface{}{
				"message": messages.InternalServerErrorMsg,
			},
			input: dto.LoginRequest{
				Email:    "lionel@gmail.com",
				Password: "testPassword",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockAuthService,
			) {
				mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(
					nil,
					assert.AnError,
				)
			},
		},
		{
			name:       "incorrect email or password",
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"message": InvalidCredentialsMsg,
			},
			input: dto.LoginRequest{
				Email:    "lionel@gmail.com",
				Password: "testPassword1",
			},
			ctx: func(ctx context.Context) context.Context {
				return context.Background()
			},
			mock: func(
				mockService *auth.MockAuthService,
			) {
				mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(
					nil,
					user.ErrInvalidCredentials,
				)
			},
		},
	}
	// Run your test cases
	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				mockService := auth.NewMockAuthService(ctrl)

				tc.mock(mockService)

				body, err := json.Marshal(tc.input)
				require.NoError(t, err)

				rctx := chi.NewRouteContext()

				req, err := http.NewRequestWithContext(
					context.WithValue(tc.ctx(context.Background()), chi.RouteCtxKey, rctx),
					http.MethodPost,
					"/api/v1/auth/login",
					bytes.NewReader(body),
				)
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				h := NewAuthHandler(mockLog, mockService)
				h.Login().ServeHTTP(recorder, req)

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

func TestRefreshHandler(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLog := zaptest.NewLogger(t)

	testCases := []struct {
		name  string
		ctx   func(ctx context.Context) context.Context
		input dto.RefreshRequest
		mock  func(
			mockService *auth.MockAuthService,
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
				mockService *auth.MockAuthService,
			) {
				mockService.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Return(
					&domain.AuthTokens{
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
				mockService *auth.MockAuthService,
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
				mockService *auth.MockAuthService,
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
				mockService *auth.MockAuthService,
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
				mockService := auth.NewMockAuthService(ctrl)

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
		mock       func(mockService *auth.MockAuthService)
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
			mock: func(mockService *auth.MockAuthService) {
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
			mock: func(mockService *auth.MockAuthService) {
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
			mock: func(mockService *auth.MockAuthService) {
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
			mock: func(mockService *auth.MockAuthService) {
				mockService.EXPECT().
					RevokeRefreshToken(gomock.Any(), gomock.Any()).
					Return(auth.ErrRefreshTokenAlreadyRevoked)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := auth.NewMockAuthService(ctrl)
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
