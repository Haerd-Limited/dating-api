package domain

import (
	"time"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

// Result states returned to the API layer.
const (
	StateRevealed        = "revealed"
	StateGatedMatchLimit = "gated_match_limit"
	StateAwaitingNext    = "awaiting_next"
	StateNone            = "none"
)

// Persisted batch states, matching the daily_pick_batches CHECK constraint.
const (
	BatchPending   = "pending"
	BatchRevealed  = "revealed"
	BatchCompleted = "completed"
)

// Pick statuses, matching the daily_picks CHECK constraint.
const (
	PickPending = "pending"
	PickLiked   = "liked"
	PickPassed  = "passed"
	PickExpired = "expired"
)

const MatchSlotLimit int64 = 2

type Batch struct {
	ID         string
	UserID     string
	BatchDate  time.Time
	State      string
	RevealedAt *time.Time
	NotifiedAt *time.Time
	Picks      []Pick
}

type Pick struct {
	ID                   string
	BatchID              string
	UserID               string
	TargetUserID         string
	Position             int
	CompatibilityPercent int
	OverlapCount         int
	PrefsSatisfied       int
	Status               string
	DecidedAt            *time.Time
}

type ScoredCandidate struct {
	UserID               string
	CompatibilityPercent int
	OverlapCount         int
	PrefsSatisfied       int
}

type DailyPicksResult struct {
	State              string
	Picks              []profilecard.ProfileCard
	ActiveMatchesCount int64
	MatchSlotLimit     int64
	UndecidedCount     int
	NextDropAt         *time.Time
}

type GenerationStats struct {
	UsersProcessed int
	BatchesCreated int
	Skipped        int
	Notified       int
}

type ResolveResult struct {
	State         string
	Batch         *Batch
	NewlyRevealed bool
}
