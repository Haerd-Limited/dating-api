package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

func TestMapToGetLikesResponseBucketsFreeAndFull(t *testing.T) {
	likes := &domain.Likes{
		FreeToMatch: []domain.Like{
			{
				Profile:            profilecard.ProfileCard{UserID: "free-user"},
				TargetAtMatchLimit: false,
				IsFavourited:       true,
			},
		},
		SlotsFull: []domain.Like{
			{
				Profile:            profilecard.ProfileCard{UserID: "full-user"},
				TargetAtMatchLimit: true,
			},
		},
		ActiveMatchesCount: 1,
		MatchSlotLimit:     constants.MaxActiveMatches,
	}

	resp := MapToGetLikesResponse(likes)

	require.Len(t, resp.FreeToMatch, 1)
	require.Len(t, resp.SlotsFull, 1)
	assert.Equal(t, "free-user", resp.FreeToMatch[0].Profile.UserID)
	assert.True(t, resp.FreeToMatch[0].IsFavourited)
	assert.False(t, resp.FreeToMatch[0].TargetAtMatchLimit)
	assert.Equal(t, "full-user", resp.SlotsFull[0].Profile.UserID)
	assert.True(t, resp.SlotsFull[0].TargetAtMatchLimit)
	assert.Equal(t, int64(1), resp.ActiveMatchesCount)
	assert.Equal(t, constants.MaxActiveMatches, resp.MatchSlotLimit)
}
