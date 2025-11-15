package middleware

import (
	"crypto/subtle"
	"net/http"
)

const adminTokenHeader = "X-Admin-Token"

func AdminMiddleware(adminAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if adminAPIKey == "" {
				http.Error(w, "admin access not configured", http.StatusForbidden)
				return
			}

			token := r.Header.Get(adminTokenHeader)
			if subtle.ConstantTimeCompare([]byte(token), []byte(adminAPIKey)) != 1 {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
