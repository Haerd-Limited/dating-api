package dto

import (
	"time"

	profiledto "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

type GetDailyPicksResponse struct {
	State              string                   `json:"state"`
	Picks              []profiledto.ProfileCard `json:"picks"`
	ActiveMatchesCount int64                    `json:"active_matches_count"`
	MatchSlotLimit     int64                    `json:"match_slot_limit"`
	UndecidedCount     int                      `json:"undecided_count"`
	NextDropAt         *time.Time               `json:"next_drop_at,omitempty"`
}
