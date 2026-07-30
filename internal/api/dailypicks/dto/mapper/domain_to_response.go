package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/dailypicks/dto"
	"github.com/Haerd-Limited/dating-api/internal/dailypicks/domain"
	dto2 "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

func DomainToGetDailyPicksResponse(result domain.DailyPicksResult) dto.GetDailyPicksResponse {
	picks := dto2.ProfileCardsToDto(result.Picks)
	if picks == nil {
		picks = []dto2.ProfileCard{}
	}

	return dto.GetDailyPicksResponse{
		State:              result.State,
		Picks:              picks,
		ActiveMatchesCount: result.ActiveMatchesCount,
		MatchSlotLimit:     result.MatchSlotLimit,
		UndecidedCount:     result.UndecidedCount,
		NextDropAt:         result.NextDropAt,
	}
}
