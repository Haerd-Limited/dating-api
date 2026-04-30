package interaction

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

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

// TestCreateHandlerMatchCapErrors locks in the HTTP contract for the new
// HAE-411 cap errors: both must surface as 409 with their dedicated user-
// facing messages so the FE can prompt the user to unmatch.
func TestCreateHandlerMatchCapErrors(t *testing.T) {
	const userID = "viewer-1"

	mockLog := zaptest.NewLogger(t)

	cases := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "actor at match limit returns 409",
			serviceErr: interaction.ErrMatchLimitReached,
			wantStatus: http.StatusConflict,
			wantMsg:    "You already have 2 active matches. Unmatch someone to make room.",
		},
		{
			name:       "target at match limit returns 409",
			serviceErr: interaction.ErrTargetMatchLimitReached,
			wantStatus: http.StatusConflict,
			wantMsg:    "This person already has the maximum number of active matches.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := interaction.NewMockService(ctrl)

			mockService.EXPECT().
				CreateSwipe(gomock.Any(), gomock.Any()).
				Return("", tc.serviceErr)

			body, err := json.Marshal(dto.SwipesRequest{
				TargetUserID: "target-2",
				Action:       "like",
			})
			require.NoError(t, err)

			ctx := context.WithValue(context.Background(), commoncontext.UserIDKey, userID)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, chi.NewRouteContext())

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/swipes", bytes.NewReader(body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			h := NewInteractionHandler(mockLog, mockService)
			h.Create().ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantStatus, recorder.Code)

			var actual map[string]any

			err = json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.wantMsg, actual["error"])

			require.True(t, ctrl.Satisfied(), "mock expectations not satisfied")
		})
	}
}
