package dto

import (
	"time"

	profiledto "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

type GetVoicesWorthHearingResponse struct {
	Profiles []profiledto.ProfileCard `json:"profiles"`
}

type GetDiscoverResponse struct {
	Profiles []profiledto.ProfileCard `json:"profiles"`
	Quota    DiscoverQuota            `json:"quota"`
}

type DiscoverQuota struct {
	Limit                int        `json:"limit"`
	WindowSeconds        int64      `json:"windowSeconds"`
	SwipesUsed           int        `json:"swipesUsed"`
	SwipesRemaining      int        `json:"swipesRemaining"`
	NextBatchAvailableAt *time.Time `json:"nextBatchAvailableAt,omitempty"`
	Exhausted            bool       `json:"exhausted"`
}
