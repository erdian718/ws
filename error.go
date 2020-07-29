package ws

import (
	"fmt"
	"net/http"
)

// Error is the ws error.
type Error struct {
	code int
	text string
}

// Error returns the error string.
func (a *Error) Error() string {
	return "ws: " + a.text
}

// IsMissingParam checks if the error is missing param.
func IsMissingParam(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.code == 0
	}
	return false
}

// MissingParam creates a missing param error.
func MissingParam(v interface{}) error {
	return &Error{
		text: fmt.Sprintf("missing param: %v", v),
	}
}

// BadRequest creates a bad request error.
func BadRequest(v interface{}) error {
	return &Error{
		code: http.StatusBadRequest,
		text: fmt.Sprintf("bad request: %v", v),
	}
}

// NotFound creates a not found error.
func NotFound(v interface{}) error {
	return &Error{
		code: http.StatusNotFound,
		text: fmt.Sprintf("not found: %v", v),
	}
}

// MethodNotAllowed creates a method not allowed error.
func MethodNotAllowed(v interface{}) error {
	return &Error{
		code: http.StatusMethodNotAllowed,
		text: fmt.Sprintf("method not allowed: %v", v),
	}
}
