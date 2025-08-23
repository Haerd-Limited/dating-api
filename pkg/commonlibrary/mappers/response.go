package mappers

type SimpleErrorResponse struct {
	Error string `json:"error"`
}

func ToSimpleErrorResponse(errMsg string) *SimpleErrorResponse {
	return &SimpleErrorResponse{
		Error: errMsg,
	}
}
