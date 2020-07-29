package ws

import "net/http"

// Error is the ws error.
type Error struct {
	code int
	text string
}

// Error returns the error string.
func (a *Error) Error() string {
	return "ws: " + a.text
}

// IsMissingParamError checks if the error is missing param.
func IsMissingParamError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.code == 0
	}
	return false
}

// MissingParamError creates a missing param error.
func MissingParamError(text string) error {
	return &Error{
		text: "missing param: " + text,
	}
}

// BadRequestError creates a bad request error.
func BadRequestError(text string) error {
	return &Error{
		code: http.StatusBadRequest,
		text: "bad request: " + text,
	}
}

// NotFoundError creates a not found error.
func NotFoundError(text string) error {
	return &Error{
		code: http.StatusNotFound,
		text: "not found: " + text,
	}
}

// MethodNotAllowedError creates a method not allowed error.
func MethodNotAllowedError(text string) error {
	return &Error{
		code: http.StatusMethodNotAllowed,
		text: "method not allowed: " + text,
	}
}

// InternalServerError creates a internal server error.
func InternalServerError(text string) error {
	return &Error{
		code: http.StatusInternalServerError,
		text: "internal server error: " + text,
	}
}
