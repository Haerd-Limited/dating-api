package profile

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

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/preference"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

func TestSetAnalyticsOptOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefSvc := preference.NewMockService(ctrl)
	h := NewProfileHandler(zaptest.NewLogger(t), nil, nil, nil, nil, prefSvc)

	cases := []struct {
		name     string
		optedOut bool
	}{
		{name: "opt out true", optedOut: true},
		{name: "opt in false", optedOut: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(dto.AnalyticsOptOutRequest{OptedOut: tc.optedOut})
			require.NoError(t, err)

			prefSvc.EXPECT().SetAnalyticsOptOut(gomock.Any(), "user-1", tc.optedOut).Return(nil)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/preferences/analytics-opt-out", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1"))
			rec := httptest.NewRecorder()
			h.SetAnalyticsOptOut().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var resp dto.AnalyticsOptOutResponse

			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tc.optedOut, resp.OptedOut)
		})
	}
}
