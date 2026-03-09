package dto

import (
	"fmt"

	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
)

type GetDiscoverRequest struct {
	Limit   int                     `json:"limit" form:"limit"`
	Offset  int                     `json:"offset" form:"offset"`
	Filters *domain.DiscoverFilters `json:"filters,omitempty" form:"filters"`
}

// Validate implements the Validator interface
func (r *GetDiscoverRequest) Validate() error {
	if r.Filters != nil && r.Filters.SeekGender != nil {
		s := *r.Filters.SeekGender
		if s != domain.SeekGenderMale && s != domain.SeekGenderFemale && s != domain.SeekGenderBoth {
			return fmt.Errorf("seek_gender must be %q, %q, or %q", domain.SeekGenderMale, domain.SeekGenderFemale, domain.SeekGenderBoth)
		}
	}
	return nil
}
