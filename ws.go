package ws

import (
	"fmt"
	"net/http"
	"strings"
)

// StatusError is the ws status error.
type StatusError struct {
	code         int
	text         string
	isHTTPStatus bool
}

// Status creates a new ws status error.
func Status(code int, msg interface{}) *StatusError {
	isHTTPStatus, tfmt := false, "unknown error: %v"
	if code == 0 {
		tfmt = "missing param: %v"
	} else if text := strings.ToLower(http.StatusText(code)); text != "" {
		isHTTPStatus = true
		tfmt = text + ": %v"
	}
	return &StatusError{
		isHTTPStatus: isHTTPStatus,
		code:         code,
		text:         fmt.Sprintf(tfmt, msg),
	}
}

// Error returns the error string.
func (a *StatusError) Error() string {
	return "ws: " + a.text
}

// Code returns the error code.
func (a *StatusError) Code() int {
	return a.code
}
