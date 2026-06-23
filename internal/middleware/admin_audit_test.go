package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	internalauditlog "github.com/Haerd-Limited/dating-api/internal/auditlog"
	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
)

func TestAdminAudit_RecordsRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	svc := internalauditlog.NewMockService(ctrl)
	logger := zaptest.NewLogger(t)

	const adminToken = "super-secret-admin-key"

	svc.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, entry domain.Entry) error {
			assert.Equal(t, http.MethodGet, entry.Method)
			assert.Equal(t, "/admin/reports/report-42", entry.Path)
			require.NotNil(t, entry.ActorIP)
			assert.Equal(t, "203.0.113.5", *entry.ActorIP)
			assert.Equal(t, tokenFingerprint(adminToken), entry.TokenFP)
			assert.NotEqual(t, adminToken, entry.TokenFP)
			require.NotNil(t, entry.TargetID)
			assert.Equal(t, "report-42", *entry.TargetID)
			assert.Equal(t, http.StatusOK, entry.StatusCode)

			return nil
		},
	)

	inner := AdminAudit(svc, logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := chi.NewRouter()
	r.Get("/admin/reports/{reportID}", inner.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/admin/reports/report-42", nil)
	req.Header.Set(adminTokenHeader, adminToken)
	req.RemoteAddr = "203.0.113.5:12345"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTokenFingerprint_IsStableHashPrefix(t *testing.T) {
	t.Parallel()

	fp1 := tokenFingerprint("same-token")
	fp2 := tokenFingerprint("same-token")
	fp3 := tokenFingerprint("other-token")

	assert.Equal(t, fp1, fp2)
	assert.NotEqual(t, fp1, fp3)
	assert.Len(t, fp1, 16)
}
