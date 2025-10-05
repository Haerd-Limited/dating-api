package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
	dto2 "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

func MapToGetVoicesWorthHearingResponse(models []profilecard.ProfileCard) dto.GetVoicesWorthHearingResponse {
	return dto.GetVoicesWorthHearingResponse{
		Profiles: dto2.ProfileCardsToDto(models),
	}
}

func MapToGetDiscoverResponse(models []profilecard.ProfileCard) dto.GetDiscoverResponse {
	return dto.GetDiscoverResponse{
		Profiles: dto2.ProfileCardsToDto(models),
	}
}
