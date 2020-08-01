// Package ws provides HTTP server implementations.
package ws

import (
	"net/http"
	"strconv"
)

// StatusError is the http status error.
type StatusError struct {
	code int
	text string
}

// Status creates a new http status error.
func Status(code int, text string) *StatusError {
	return &StatusError{
		code: code,
		text: text,
	}
}

// Code returns the error code.
func (a *StatusError) Code() int {
	return a.code
}

// Error returns the error string.
func (a *StatusError) Error() string {
	return "ws: " + strconv.Itoa(a.code) + " " + a.text
}

func finally(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := http.StatusInternalServerError
	if e, ok := err.(*StatusError); ok {
		code = e.code
	}
	text := http.StatusText(code)
	if text == "" {
		code = http.StatusTeapot
		text = http.StatusText(code)
	}
	http.Error(w, text, code)
}
