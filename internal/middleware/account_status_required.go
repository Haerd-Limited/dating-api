package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

type accountBannedResponse struct {
	Error string `json:"error"`
}

type accountSuspendedResponse struct {
	Error string `json:"error"`
	Until string `json:"until"`
}

type warningAckRequiredResponse struct {
	Error string `json:"error"`
}

func AccountStatusRequired(userService user.Service, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, ok := commoncontext.UserIDFromContext(ctx)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			state, err := userService.GetAccountGateState(ctx, userID)
			if err != nil {
				logger.Warn("account status gate lookup failed", zap.String("userID", userID), zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			now := time.Now().UTC()
			effective := state.EffectiveStatus(now)

			switch effective {
			case userdomain.AccountStatusBanned:
				writeAccountStatusJSON(w, http.StatusForbidden, accountBannedResponse{Error: "account_banned"})
				return
			case userdomain.AccountStatusSuspended:
				until := ""
				if state.SuspendedUntil != nil {
					until = state.SuspendedUntil.UTC().Format(time.RFC3339)
				}

				writeAccountStatusJSON(w, http.StatusForbidden, accountSuspendedResponse{
					Error: "account_suspended",
					Until: until,
				})

				return
			}

			if state.HasPendingWarning {
				writeAccountStatusJSON(w, http.StatusForbidden, warningAckRequiredResponse{
					Error: "warning_acknowledgement_required",
				})

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAccountStatusJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
