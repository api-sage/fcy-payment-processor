package commons

type Response[T any] struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    *T       `json:"data,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func SuccessResponse[T any](message string, data T) Response[T] {
	return Response[T]{
		Success: true,
		Message: message,
		Data:    &data,
	}
}

func ErrorResponse[T any](message string, errors ...string) Response[T] {
	return Response[T]{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
