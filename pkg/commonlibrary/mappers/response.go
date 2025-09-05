package mappers

type SimpleErrorResponse struct {
	Error string `json:"error"`
}

func ToSimpleErrorResponse(errMsg string) *SimpleErrorResponse {
	return &SimpleErrorResponse{
		Error: errMsg,
	}
}

type SimpleMessageResponse struct {
	Message string `json:"message"`
}

func ToSimpleMessageResponse(msg string) *SimpleMessageResponse {
	return &SimpleMessageResponse{
		Message: msg,
	}
}
