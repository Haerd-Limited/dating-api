package discover

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

func TestNewDiscoverFeedResultAtMatchLimit(t *testing.T) {
	quota := domain.NewQuotaStatus(domain.DiscoverQuotaLimit, domain.DiscoverQuotaWindow, 0, nil)
	result := domain.NewDiscoverFeedResultAtMatchLimit(quota)

	assert.True(t, result.AtMatchLimit)
	assert.Equal(t, constants.MaxActiveMatches, result.MatchSlotLimit)
	assert.Nil(t, result.Profiles)
}
