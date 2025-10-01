package dto

type UploadUrl struct {
	Key       string            `json:"key"`
	UploadUrl string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
	MaxBytes  int64             `json:"max_bytes"`
}
