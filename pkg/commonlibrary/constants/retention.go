package constants

import "time"

const (
	RetentionVerificationCodes         = 30 * 24 * time.Hour
	RetentionEvents                    = 2 * 365 * 24 * time.Hour
	RetentionVerificationAttempts      = 1 * 365 * 24 * time.Hour
	RetentionExpiredRefreshTokensGrace = 7 * 24 * time.Hour
	RetentionPurgeBatchSize            = 10000
)
