package render

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/communication"
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

// ErrorNotificationHandler handles sending SMS notifications for errors
type ErrorNotificationHandler struct {
	communicationService         communication.Service
	backendEngineerPhoneNumbers  []string
	frontendEngineerPhoneNumbers []string
	logger                       *zap.Logger
}

// NewErrorNotificationHandler creates a new error notification handler
func NewErrorNotificationHandler(
	communicationService communication.Service,
	backendEngineerPhoneNumbers []string,
	frontendEngineerPhoneNumbers []string,
	logger *zap.Logger,
) *ErrorNotificationHandler {
	return &ErrorNotificationHandler{
		communicationService:         communicationService,
		backendEngineerPhoneNumbers:  backendEngineerPhoneNumbers,
		frontendEngineerPhoneNumbers: frontendEngineerPhoneNumbers,
		logger:                       logger,
	}
}

// sendErrorNotification sends SMS notifications to appropriate engineers based on status code
func (enh *ErrorNotificationHandler) sendErrorNotification(handlerName string, statusCode int, errMsg string, err error) {
	if enh.communicationService == nil {
		return // No communication service configured
	}

	var phoneNumbers []string

	var recipientType string

	if statusCode == http.StatusInternalServerError {
		phoneNumbers = enh.backendEngineerPhoneNumbers
		recipientType = "Backend Engineers"
	} else if statusCode >= 400 && statusCode < 500 {
		phoneNumbers = enh.frontendEngineerPhoneNumbers
		recipientType = "Frontend Engineers"
	} else {
		return // Don't send notifications for other status codes
	}

	if len(phoneNumbers) == 0 {
		return // No phone numbers configured
	}

	// Build SMS message
	errorDetails := errMsg
	if err != nil && err.Error() != errMsg {
		errorDetails = fmt.Sprintf("%s (Error: %s)", errMsg, err.Error())
	}

	message := fmt.Sprintf("API Error Alert\nHandler: %s\nStatus: %d\nMessage: %s", handlerName, statusCode, errorDetails)

	// Send SMS to all configured phone numbers
	for _, phoneNumber := range phoneNumbers {
		if phoneNumber == "" {
			continue
		}

		smsErr := enh.communicationService.SendSMS(phoneNumber, message)
		if smsErr != nil {
			enh.logger.Sugar().Errorw("failed to send error notification SMS",
				"recipientType", recipientType,
				"phoneNumber", phoneNumber,
				"handlerName", handlerName,
				"statusCode", statusCode,
				"error", smsErr)
		}
	}
}

var globalErrorNotificationHandler *ErrorNotificationHandler

// InitErrorNotificationHandler initializes the global error notification handler
func InitErrorNotificationHandler(
	communicationService communication.Service,
	backendEngineerPhoneNumbers []string,
	frontendEngineerPhoneNumbers []string,
	logger *zap.Logger,
) {
	globalErrorNotificationHandler = NewErrorNotificationHandler(
		communicationService,
		backendEngineerPhoneNumbers,
		frontendEngineerPhoneNumbers,
		logger,
	)
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

	// Send SMS notifications for 500 and 4XX errors
	if globalErrorNotificationHandler != nil {
		globalErrorNotificationHandler.sendErrorNotification(handlerName, statusCode, errMsg, err)
	}

	Json(w, statusCode, mappers.ToSimpleErrorResponse(errMsg))
}
