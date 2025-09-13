package domain

type Swipe struct {
	TargetUserID   string
	Action         string
	UserID         string
	IdempotencyKey *string
}

type Match struct {
	UserID         string
	DisplayName    string
	MessagePreview string
	Emoji          string
	Reveal         bool
	RevealProgress int
}
