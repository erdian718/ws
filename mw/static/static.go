package static

import (
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
		return ctx.Next()
	}
}
