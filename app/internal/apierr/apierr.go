package apierr

import "net/http"

// Detail is the contract for structured API errors.
type Detail interface {
	error
	HTTPStatus() int
	Trace() interface{}
}

type apiError struct {
	message string
	status  int
	trace   interface{}
}

func (e *apiError) Error() string      { return e.message }
func (e *apiError) HTTPStatus() int    { return e.status }
func (e *apiError) Trace() interface{} { return e.trace }

// New creates a 400 Bad Request error.
func New(message string, trace interface{}) Detail {
	return &apiError{message: message, status: http.StatusBadRequest, trace: trace}
}

// NewWithStatus creates an error with a specific HTTP status code.
func NewWithStatus(message string, trace interface{}, status int) Detail {
	return &apiError{message: message, status: status, trace: trace}
}

// NotFound creates a 404 Not Found error.
func NotFound(message string) Detail {
	return &apiError{message: message, status: http.StatusNotFound}
}

// Conflict creates a 409 Conflict error.
func Conflict(message string) Detail {
	return &apiError{message: message, status: http.StatusConflict}
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string) Detail {
	return &apiError{message: message, status: http.StatusUnauthorized}
}

// BadGateway creates a 502 Bad Gateway error.
func BadGateway(message string, trace interface{}) Detail {
	return &apiError{message: message, status: http.StatusBadGateway, trace: trace}
}

// UnprocessableEntity creates a 422 Unprocessable Entity error.
func UnprocessableEntity(message string, trace interface{}) Detail {
	return &apiError{message: message, status: http.StatusUnprocessableEntity, trace: trace}
}
