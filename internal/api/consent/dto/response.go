package dto

type ConsentRequiredErrorResponse struct {
	Error   string   `json:"error"`
	Missing []string `json:"missing"`
}
