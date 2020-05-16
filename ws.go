package ws

import (
	"errors"
	"net/http"
)

// Errors.
var (
	ErrMissingParam     = errors.New("ws: missing parameter")
	ErrBadRequest       = errors.New(http.StatusText(http.StatusBadRequest))
	ErrNotFound         = errors.New(http.StatusText(http.StatusNotFound))
	ErrMethodNotAllowed = errors.New(http.StatusText(http.StatusMethodNotAllowed))
)
