package logger

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// LogError logs an error with structured fields and returns a clean error message.
// This helper ensures consistent error logging across all services by:
// - Logging errors with structured fields (zap) for better searchability
// - Returning clean error messages without embedded variable values
// - Accepting optional context fields (userID, conversationID, etc.)
//
// Context cancellations and deadline exceeded errors are not genuine server
// failures (the client closed the connection or the request timed out), so they
// are logged at a lower severity to avoid polluting error logs and triggering
// alerts. The wrapped error is still returned so callers/handlers behave the same.
func LogError(logger *zap.Logger, operation string, err error, fields ...zap.Field) error {
	allFields := append([]zap.Field{zap.Error(err), zap.String("operation", operation)}, fields...)

	switch {
	case errors.Is(err, context.Canceled):
		logger.Info(operation, allFields...)
	case errors.Is(err, context.DeadlineExceeded):
		logger.Warn(operation, allFields...)
	default:
		logger.Error(operation, allFields...)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
