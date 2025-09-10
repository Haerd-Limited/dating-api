package domain

type Swipe struct {
	TargetUserID   string
	Action         string
	UserID         string
	IdempotencyKey *string
}
