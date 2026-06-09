package domain

import (
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

type DiscoverFeedResult struct {
	Profiles       []profilecard.ProfileCard
	Quota          QuotaStatus
	AtMatchLimit   bool
	MatchSlotLimit int64
}

func NewDiscoverFeedResult(profiles []profilecard.ProfileCard, quota QuotaStatus) DiscoverFeedResult {
	return DiscoverFeedResult{
		Profiles:       profiles,
		Quota:          quota,
		MatchSlotLimit: constants.MaxActiveMatches,
	}
}

func NewDiscoverFeedResultAtMatchLimit(quota QuotaStatus) DiscoverFeedResult {
	return DiscoverFeedResult{
		Profiles:       nil,
		Quota:          quota,
		AtMatchLimit:   true,
		MatchSlotLimit: constants.MaxActiveMatches,
	}
}
