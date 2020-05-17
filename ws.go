package ws

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Errors.
var (
	ErrMissingParam     = errors.New("ws: missing parameter")
	ErrBadRequest       = errors.New(http.StatusText(http.StatusBadRequest))
	ErrNotFound         = errors.New(http.StatusText(http.StatusNotFound))
	ErrMethodNotAllowed = errors.New(http.StatusText(http.StatusMethodNotAllowed))
)

func sendFile(w http.ResponseWriter, r *http.Request, name string) error {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	http.ServeContent(w, r, f.Name(), stat.ModTime(), f)
	return nil
}

func handleError(w http.ResponseWriter, r *http.Request, err error, h func(*http.Request, error)) {
	if err != nil {
		if errors.Is(err, ErrMissingParam) {
			w.WriteHeader(http.StatusBadRequest)
		} else if errors.Is(err, ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
		} else if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, ErrMethodNotAllowed) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if h != nil {
			h(r, err)
		}
	}
}
