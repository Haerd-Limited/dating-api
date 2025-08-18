package mappers

type SimpleMessageResponse struct {
	Message string `json:"message"`
}

func ToSimpleMessageResponse(message string) *SimpleMessageResponse {
	return &SimpleMessageResponse{
		Message: message,
	}
}
