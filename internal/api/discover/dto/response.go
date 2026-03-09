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

type GetUserPreferencesResponse struct {
	DistanceKM         *int    `json:"distance_km,omitempty"`
	MinAge             *int    `json:"min_age,omitempty"`
	MaxAge             *int    `json:"max_age,omitempty"`
	DatingIntentionIDs []int16 `json:"dating_intention_ids,omitempty"`
	ReligionIDs        []int16 `json:"religion_ids,omitempty"`
	SexualityIDs       []int16 `json:"sexuality_ids,omitempty"`
	EthnicityIDs       []int16 `json:"ethnicity_ids,omitempty"`
	SeekGenderIDs      []int16 `json:"seek_gender_ids,omitempty"`
	SeekGender         string  `json:"seek_gender,omitempty"` // "Male", "Female", or "Both"
}
