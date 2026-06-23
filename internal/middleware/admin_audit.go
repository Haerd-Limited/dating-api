package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	internalauditlog "github.com/Haerd-Limited/dating-api/internal/auditlog"
	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// AdminAudit logs every request under /admin for breach-reconstruction (GDPR Art. 33).
func AdminAudit(svc internalauditlog.Service, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			entry := domain.Entry{
				OccurredAt: time.Now().UTC(),
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rec.status,
			}

			if ip := request.ClientIP(r); ip != "" {
				entry.ActorIP = &ip
			}

			if token := r.Header.Get(adminTokenHeader); token != "" {
				entry.TokenFP = tokenFingerprint(token)
			}

			if targetID := adminTargetID(r); targetID != "" {
				entry.TargetID = &targetID
			}

			if sessionID, displayName, ok := context.AdminActorFromContext(r.Context()); ok {
				entry.ActorSessionID = &sessionID
				if displayName != "" {
					entry.ActorName = &displayName
				}
			}

			if err := svc.Record(r.Context(), entry); err != nil {
				logger.Warn("admin audit log record failed",
					zap.String("method", entry.Method),
					zap.String("path", entry.Path),
					zap.Error(err))
			}
		})
	}
}

func tokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}

func adminTargetID(r *http.Request) string {
	for _, key := range []string{"reportID", "attemptID", "userID"} {
		if id := chi.URLParam(r, key); id != "" {
			return id
		}
	}

	return ""
}
