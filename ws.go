// Package ws provides HTTP server implementations.
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
	tfmt := http.StatusText(code)
	if tfmt == "" {
		code = http.StatusTeapot
		tfmt = http.StatusText(code)
	}
	return &StatusError{
		code: code,
		text: fmt.Sprintf(strings.ToLower(tfmt)+": %v", msg),
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

func finally(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := http.StatusInternalServerError
	if e, ok := err.(*StatusError); ok {
		code = e.code
	}
	http.Error(w, http.StatusText(code), code)
}
