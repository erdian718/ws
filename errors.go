package ws

import (
	"/net/http"
	"errors"
)

// Errors.
var (
	ErrMissingParam = errors.New("ws: missing parameter")
	ErrBadRequest   = errors.New(http.StatusText(http.StatusBadRequest))
	ErrNotFound     = errors.New(http.StatusText(http.StatusNotFound))
)
