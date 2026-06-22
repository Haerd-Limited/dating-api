package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	internalconsent "github.com/Haerd-Limited/dating-api/internal/consent"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

type consentRequiredResponse struct {
	Error   string   `json:"error"`
	Missing []string `json:"missing"`
}

func ConsentRequired(svc internalconsent.Service, enabled bool, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			userID, ok := commoncontext.UserIDFromContext(ctx)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			missing, err := svc.GetMissingMandatory(ctx, userID)
			if err != nil {
				logger.Warn("consent gate lookup failed", zap.String("userID", userID), zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			if len(missing) > 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(consentRequiredResponse{
					Error:   "consent_required",
					Missing: missing,
				})

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
