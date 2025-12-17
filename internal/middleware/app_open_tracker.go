package middleware

import (
	"net/http"

	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

// AppOpenTracker middleware automatically tracks app opens for authenticated users.
// It should be placed after AuthMiddleware so that userID is available in context.
func AppOpenTracker() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract userID from context (set by AuthMiddleware)
			userID, ok := commoncontext.UserIDFromContext(ctx)
			if ok {
				// Track app open event asynchronously
				// Analytics is already non-blocking, so this won't slow down requests
				props := make(map[string]any)
				if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
					props["user_agent"] = userAgent
				}

				if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
					props["session_id"] = sessionID
				}

				commonanalytics.Track(ctx, "app.opened", &userID, nil, props)
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
