package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// LogError logs an error with structured fields and returns a clean error message.
// This helper ensures consistent error logging across all services by:
// - Logging errors with structured fields (zap) for better searchability
// - Returning clean error messages without embedded variable values
// - Accepting optional context fields (userID, conversationID, etc.)
func LogError(logger *zap.Logger, operation string, err error, fields ...zap.Field) error {
	allFields := append([]zap.Field{zap.Error(err), zap.String("operation", operation)}, fields...)
	logger.Error(operation, allFields...)

	return fmt.Errorf("%s: %w", operation, err)
}
