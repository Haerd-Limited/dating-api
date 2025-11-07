package domain

import "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"

type DiscoverFeedResult struct {
	Profiles []profilecard.ProfileCard
	Quota    QuotaStatus
}

func NewDiscoverFeedResult(profiles []profilecard.ProfileCard, quota QuotaStatus) DiscoverFeedResult {
	return DiscoverFeedResult{
		Profiles: profiles,
		Quota:    quota,
	}
}
