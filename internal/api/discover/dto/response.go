package dto

import "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"

type GetVoicesWorthHearingResponse struct {
	Profiles []dto.ProfileCard `json:"profiles"`
}

type GetDiscoverResponse struct {
	Profiles []dto.ProfileCard `json:"profiles"`
}
