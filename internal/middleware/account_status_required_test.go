package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

func TestAccountStatusRequired(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(24 * time.Hour)
	past := time.Now().UTC().Add(-24 * time.Hour)

	cases := []struct {
		name       string
		setupMock  func(*user.MockService)
		wantStatus int
		wantError  string
	}{
		{
			name: "active passes",
			setupMock: func(svc *user.MockService) {
				svc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
					Status: userdomain.AccountStatusActive,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "banned blocked",
			setupMock: func(svc *user.MockService) {
				svc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
					Status: userdomain.AccountStatusBanned,
				}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantError:  "account_banned",
		},
		{
			name: "suspended active blocked",
			setupMock: func(svc *user.MockService) {
				svc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
					Status:         userdomain.AccountStatusSuspended,
					SuspendedUntil: &future,
				}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantError:  "account_suspended",
		},
		{
			name: "suspended expired passes",
			setupMock: func(svc *user.MockService) {
				svc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
					Status:         userdomain.AccountStatusSuspended,
					SuspendedUntil: &past,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "pending warning blocked",
			setupMock: func(svc *user.MockService) {
				svc.EXPECT().GetAccountGateState(gomock.Any(), "user-1").Return(userdomain.AccountState{
					Status:            userdomain.AccountStatusActive,
					HasPendingWarning: true,
				}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantError:  "warning_acknowledgement_required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := user.NewMockService(ctrl)
			tc.setupMock(svc)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true

				w.WriteHeader(http.StatusOK)
			})

			handler := AccountStatusRequired(svc, zaptest.NewLogger(t))(next)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/discover", nil)
			ctx := context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req.WithContext(ctx))

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				assert.True(t, nextCalled)
				return
			}

			assert.False(t, nextCalled)

			var body map[string]any

			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tc.wantError, body["error"])
		})
	}
}
