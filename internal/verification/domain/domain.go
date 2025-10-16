package domain

type (
	VerificationType   string
	VerificationStatus string
)

const (
	// Tune these after you collect real data
	LivenessPassThreshold = 0.80 // AWS returns boolean+scores; we’ll read confidence
	FaceMatchThreshold    = 85.0 // Rekognition similarity (0–100)
)

type StartResult struct {
	SessionID string
	Region    string
}

type CompleteResult struct {
	Status        string // passed / failed / needs_review
	MatchScore    float64
	Reasons       []string
	PhotoVerified bool
}
