package dto

import "github.com/Haerd-Limited/dating-api/internal/discover/domain"

type GetDiscoverRequest struct {
	Limit   int                     `json:"limit" form:"limit"`
	Offset  int                     `json:"offset" form:"offset"`
	Filters *domain.DiscoverFilters `json:"filters,omitempty" form:"filters"`
}

// Validate implements the Validator interface
func (r *GetDiscoverRequest) Validate() error {
	// Basic validation - limit and offset are already handled by ParseQueryInt
	// Additional validation can be added here if needed
	return nil
}
