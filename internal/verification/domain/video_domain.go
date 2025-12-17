package domain

type StartVideoResult struct {
	Code      string
	UploadURL string
	UploadKey string
}

type SubmitVideoResult struct {
	Status string // "submitted"
}
