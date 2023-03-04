package util

type ExceptionResponse struct {
	Message string
}

func NewExceptionResponse(message string) *ExceptionResponse {
	return &ExceptionResponse{message}
}

func (response *ExceptionResponse) setMessage(message string) {
	response.Message = message
}
