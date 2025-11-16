package render

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
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
	Json(w, http.StatusUnauthorized, mappers.ToSimpleErrorResponse(messages.AuthenticationRequiredMsg))
}

func ErrorCausedByTimeoutOrClientCancellation(w http.ResponseWriter, r *http.Request, logger *zap.Logger, err error) bool {
	switch {
	case errors.Is(err, context.Canceled):
		logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
		return true // no need to return a response. Client socket is closed.
	case errors.Is(err, context.DeadlineExceeded):
		Json(w, http.StatusGatewayTimeout, mappers.ToSimpleErrorResponse("request timed out"))
		return true
	default:
		return false
	}
}

func HandleServiceErrorResponse(
	logger *zap.Logger,
	w http.ResponseWriter,
	r *http.Request,
	handlerName string,
	err error,
	errorsToStatusCodeAndMessageMapper func(err error) (int, string),
) {
	if ErrorCausedByTimeoutOrClientCancellation(w, r, logger, err) {
		return
	}

	statusCode, errMsg := errorsToStatusCodeAndMessageMapper(err)

	switch {
	case statusCode == http.StatusInternalServerError:
		logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	case statusCode >= 400:
		logger.Sugar().Warnw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	default:
		logger.Sugar().Infow(fmt.Sprintf("%s response", handlerName), "message", errMsg)
	}

	Json(w, statusCode, mappers.ToSimpleErrorResponse(errMsg))
}
