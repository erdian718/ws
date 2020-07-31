package ws

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Error is the ws error.
type Error struct {
	code int
	text string
}

// NewError creates a new ws error.
func NewError(code int, msg interface{}) *Error {
	tfmt := "missing param: %v"
	if code > 0 {
		if text := strings.ToLower(http.StatusText(code)); text == "" {
			tfmt = "unknown error: %v"
		} else {
			tfmt = text + ": %v"
		}
	}
	return &Error{
		code: code,
		text: fmt.Sprintf(tfmt, msg),
	}
}

// Error returns the error string.
func (a *Error) Error() string {
	return "ws: " + a.text
}

// Code returns the error code.
func (a *Error) Code() int {
	return a.code
}

func serveFile(w http.ResponseWriter, r *http.Request, name string) error {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = NewError(http.StatusNotFound, err)
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
