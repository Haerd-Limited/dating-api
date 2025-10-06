package render

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
)

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

func Json(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set(headerContentType, contentTypeJSON)

	body, err := json.Marshal(payload)
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = []byte(fmt.Sprintf(`{"error":"%s"}`, err))
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func UnauthorizedResponse(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	authHeader := r.Header.Get("Authorization")
	logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)
	Json(w, http.StatusUnauthorized, commonMappers.ToSimpleErrorResponse(messages.AuthenticationRequiredMsg))
}

func ErrorCausedByTimeoutOrClientCancellation(w http.ResponseWriter, r *http.Request, logger *zap.Logger, err error) bool {
	switch {
	case errors.Is(err, context.Canceled):
		logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
		return true // no need to return a response. Client socket is closed.
	case errors.Is(err, context.DeadlineExceeded):
		Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
		return true
	default:
		return false
	}
}
