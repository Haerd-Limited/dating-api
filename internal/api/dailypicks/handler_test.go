package dailypicks

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

	"github.com/Haerd-Limited/dating-api/internal/api/dailypicks/dto"
	internaldailypicks "github.com/Haerd-Limited/dating-api/internal/dailypicks"
	"github.com/Haerd-Limited/dating-api/internal/dailypicks/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

const testUserID = "daily-picks-handler-user"

func TestGetDailyPicksHandler(t *testing.T) {
	mockLog := zaptest.NewLogger(t)
	nextDrop := time.Now().UTC().Add(2 * time.Hour)

	cases := []struct {
		name       string
		setupMock  func(svc *internaldailypicks.MockService)
		wantStatus int
		assertBody func(t *testing.T, body dto.GetDailyPicksResponse)
	}{
		{
			name: "revealed with two picks",
			setupMock: func(svc *internaldailypicks.MockService) {
				svc.EXPECT().GetDailyPicks(gomock.Any(), testUserID).Return(domain.DailyPicksResult{
					State:              domain.StateRevealed,
					Picks:              []profilecard.ProfileCard{{UserID: "a"}, {UserID: "b"}},
					ActiveMatchesCount: 0,
					MatchSlotLimit:     2,
					UndecidedCount:     2,
					NextDropAt:         &nextDrop,
				}, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body dto.GetDailyPicksResponse) {
				assert.Equal(t, domain.StateRevealed, body.State)
				assert.Len(t, body.Picks, 2)
				assert.Equal(t, int64(0), body.ActiveMatchesCount)
				assert.Equal(t, 2, body.UndecidedCount)
			},
		},
		{
			name: "gated match limit with empty picks",
			setupMock: func(svc *internaldailypicks.MockService) {
				svc.EXPECT().GetDailyPicks(gomock.Any(), testUserID).Return(domain.DailyPicksResult{
					State:              domain.StateGatedMatchLimit,
					Picks:              nil,
					ActiveMatchesCount: 1,
					MatchSlotLimit:     2,
					UndecidedCount:     0,
				}, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body dto.GetDailyPicksResponse) {
				assert.Equal(t, domain.StateGatedMatchLimit, body.State)
				assert.Empty(t, body.Picks)
				assert.Equal(t, int64(1), body.ActiveMatchesCount)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockSvc := internaldailypicks.NewMockService(ctrl)
			tc.setupMock(mockSvc)

			ctx := context.WithValue(context.Background(), commoncontext.UserIDKey, testUserID)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-picks", nil).WithContext(ctx)
			rec := httptest.NewRecorder()

			NewHandler(mockLog, mockSvc).GetDailyPicks().ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			var body dto.GetDailyPicksResponse

			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			tc.assertBody(t, body)
		})
	}
}
