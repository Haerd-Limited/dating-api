package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	internaladminsession "github.com/Haerd-Limited/dating-api/internal/adminsession"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

const adminSessionHeader = "X-Admin-Session"

func AdminSession(svc internaladminsession.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimSpace(r.Header.Get(adminSessionHeader))
			if token == "" {
				http.Error(w, "admin session required", http.StatusUnauthorized)
				return
			}

			session, err := svc.ValidateToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, internaladminsession.ErrSessionExpired) {
					http.Error(w, "admin session expired", http.StatusUnauthorized)
					return
				}

				http.Error(w, "invalid admin session", http.StatusUnauthorized)

				return
			}

			_ = svc.TouchSession(r.Context(), session.ID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, commoncontext.AdminSessionIDKey, session.ID)
			ctx = context.WithValue(ctx, commoncontext.AdminDisplayNameKey, session.DisplayName)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminSessionFromQuery validates a session token from ?session= (for SSE EventSource).
func AdminSessionFromQuery(svc internaladminsession.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimSpace(r.URL.Query().Get("session"))
			if token == "" {
				http.Error(w, "session query param required", http.StatusUnauthorized)
				return
			}

			session, err := svc.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid admin session", http.StatusUnauthorized)
				return
			}

			_ = svc.TouchSession(r.Context(), session.ID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, commoncontext.AdminSessionIDKey, session.ID)
			ctx = context.WithValue(ctx, commoncontext.AdminDisplayNameKey, session.DisplayName)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
