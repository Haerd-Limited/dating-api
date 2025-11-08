package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto"
	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	dto2 "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

func MapToGetVoicesWorthHearingResponse(models []profilecard.ProfileCard) dto.GetVoicesWorthHearingResponse {
	profiles := dto2.ProfileCardsToDto(models)
	if profiles == nil {
		profiles = []dto2.ProfileCard{}
	}

	return dto.GetVoicesWorthHearingResponse{
		Profiles: profiles,
	}
}

func MapToGetDiscoverResponse(result domain.DiscoverFeedResult) dto.GetDiscoverResponse {
	profiles := dto2.ProfileCardsToDto(result.Profiles)
	if profiles == nil {
		profiles = []dto2.ProfileCard{}
	}

	return dto.GetDiscoverResponse{
		Profiles: profiles,
		Quota:    mapQuota(result.Quota),
	}
}

func mapQuota(status domain.QuotaStatus) dto.DiscoverQuota {
	return dto.DiscoverQuota{
		Limit:                status.Limit,
		WindowSeconds:        int64(status.Window / time.Second),
		SwipesUsed:           status.SwipesUsed,
		SwipesRemaining:      status.SwipesRemaining,
		NextBatchAvailableAt: status.NextBatchAvailableAt,
		Exhausted:            status.Exhausted(),
	}
}
