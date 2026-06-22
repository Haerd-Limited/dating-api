package retention

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

func TestRetentionConstants(t *testing.T) {
	require.Equal(t, 10000, constants.RetentionPurgeBatchSize)
	require.Equal(t, 30*24*time.Hour, constants.RetentionVerificationCodes)
	require.Equal(t, 2*365*24*time.Hour, constants.RetentionEvents)
	require.Equal(t, 1*365*24*time.Hour, constants.RetentionVerificationAttempts)
	require.Equal(t, 7*24*time.Hour, constants.RetentionExpiredRefreshTokensGrace)
}
