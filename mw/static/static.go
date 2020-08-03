package static

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ofunc/ws"
)

var allow = strings.Join([]string{http.MethodOptions, http.MethodGet, http.MethodHead}, ", ")

// New creates a static file middleware.
func New(root string) func(*ws.Context) error {
	return func(ctx *ws.Context) error {
		err := ctx.Next()
		if err == nil {
			return err
		}
		serr, ok := err.(*ws.StatusError)
		if !ok {
			return err
		}
		if serr.Code() != http.StatusNotFound {
			return err
		}

		method := ctx.Request.Method
		if method == http.MethodOptions {
			ctx.ResponseWriter.Header().Add("Allow", allow)
			ctx.ResponseWriter.WriteHeader(http.StatusOK)
			ctx.ResponseWriter.Write([]byte(""))
			return nil
		}
		if method != http.MethodGet && method != http.MethodHead {
			return ws.Status(http.StatusMethodNotAllowed, method+" "+ctx.Request.URL.Path)
		}
		if strings.Contains(ctx.Path, "..") {
			return ws.Status(http.StatusBadRequest, "invalid path: "+ctx.Request.URL.Path)
		}
		return ctx.File(filepath.Join(root, filepath.FromSlash(ctx.Path)))
	}
}
