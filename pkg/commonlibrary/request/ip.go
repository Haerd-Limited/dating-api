package request

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP prefers X-Forwarded-For, then RemoteAddr; strips port and IPv6 brackets.
func ClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	return strings.Trim(host, "[]")
}
