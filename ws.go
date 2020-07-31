package ws

import (
	"fmt"
	"net/http"
	"strings"
)

// StatusError is the http status error.
type StatusError struct {
	code int
	text string
}

// Status creates a new http status error.
func Status(code int, msg interface{}) *StatusError {
	tfmt := "unknown error: %v"
	if text := strings.ToLower(http.StatusText(code)); text != "" {
		tfmt = text + ": %v"
	}
	return &StatusError{
		code: code,
		text: fmt.Sprintf(tfmt, msg),
	}
}

// Code returns the error code.
func (a *StatusError) Code() int {
	return a.code
}

// Error returns the error string.
func (a *StatusError) Error() string {
	return "ws: " + a.text
}
