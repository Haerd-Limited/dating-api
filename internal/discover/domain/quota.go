package domain

import (
	"time"
)

const (
	DiscoverQuotaLimit  = 10
	DiscoverQuotaWindow = 16 * time.Hour
)

type QuotaStatus struct {
	Limit                int
	Window               time.Duration
	SwipesUsed           int
	SwipesRemaining      int
	NextBatchAvailableAt *time.Time
}

func (qs QuotaStatus) Exhausted() bool {
	return qs.SwipesRemaining <= 0
}

func NewQuotaStatus(limit int, window time.Duration, swipesUsed int, gatingSwipeAt *time.Time) QuotaStatus {
	if limit < 0 {
		limit = 0
	}

	if window < 0 {
		window = 0
	}

	if swipesUsed < 0 {
		swipesUsed = 0
	}

	if swipesUsed > limit {
		swipesUsed = limit
	}

	remaining := limit - swipesUsed
	if remaining < 0 {
		remaining = 0
	}

	var next *time.Time

	if remaining == 0 && gatingSwipeAt != nil {
		nextAt := gatingSwipeAt.Add(window)
		next = &nextAt
	}

	return QuotaStatus{
		Limit:                limit,
		Window:               window,
		SwipesUsed:           swipesUsed,
		SwipesRemaining:      remaining,
		NextBatchAvailableAt: next,
	}
}
