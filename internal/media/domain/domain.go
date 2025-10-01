package domain

type UploadUrl struct {
	Key       string
	UploadUrl string
	Headers   map[string]string
	MaxBytes  int64
}
