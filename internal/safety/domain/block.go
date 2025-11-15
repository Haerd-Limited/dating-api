package domain

import "time"

type Block struct {
	BlockerID string
	BlockedID string
	Reason    *string
	CreatedAt time.Time
}

type BlockRequest struct {
	BlockerID string
	BlockedID string
	Reason    *string
}
