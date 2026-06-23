package safety

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

	dto "github.com/Haerd-Limited/dating-api/internal/api/safety/dto"
	"github.com/Haerd-Limited/dating-api/internal/safety"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

// TestAdminResolveReportSuspendMissingUntil proves the service-level validation
// error for a suspension without suspend_until surfaces as a 400.
func TestAdminResolveReportSuspendMissingUntil(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := safety.NewMockService(ctrl)

	mockService.EXPECT().
		ResolveReport(gomock.Any(), gomock.Any()).
		Return(safety.ErrInvalidSuspendUntil)

	body, err := json.Marshal(dto.ResolveReportRequest{
		ActionType: "suspend_user",
		NewStatus:  "resolved",
	})
	require.NoError(t, err)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("reportID", "report-1")

	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, commoncontext.UserIDKey, "admin-1")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/admin/reports/report-1/resolve", bytes.NewReader(body))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	h := NewHandler(zaptest.NewLogger(t), mockService)
	h.AdminResolveReport().ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var actual map[string]string

	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.Equal(t, safety.ErrInvalidSuspendUntil.Error(), actual["error"])
}
