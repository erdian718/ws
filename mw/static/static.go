package static

import (
	"net/http"
	"path/filepath"

	"github.com/ofunc/ws"
)

// New creates a static file middleware.
func New(root string) func(*ws.Context) error {
	return func(ctx *ws.Context) error {
		for i := len(ctx.Path) - 1; i >= 0 && ctx.Path[i] != '/'; i-- {
			if ctx.Path[i] == '.' {
				return ctx.File(filepath.Join(root, filepath.FromSlash(ctx.Path)))
			}
		}

		err := ctx.Next()
		if err != nil {
			if e, ok := err.(*ws.StatusError); ok {
				if e.Code() == http.StatusNotFound {
					return ctx.File(filepath.Join(root, filepath.FromSlash(ctx.Path), "index.html"))
				}
			}
		}
		return err
	}
}
