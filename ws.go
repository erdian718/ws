package ws

import (
	"net/http"
	"os"
)

func serveFile(w http.ResponseWriter, r *http.Request, name string) error {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = NotFound(err.Error())
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

func handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var code int
	if e, ok := err.(*Error); ok {
		if e.code == 0 {
			code = http.StatusBadRequest
		} else {
			code = e.code
		}
	} else {
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
}
