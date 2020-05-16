package static

import (
	"path/filepath"

	"github.com/ofunc/ws"
)

// New creates a new static middleware.
func New(root string) func(*ws.Context) error {
	return func(c *ws.Context) error {
		p, err := c.Param("*")
		if err != nil {
			return err
		}
		return c.File(filepath.Join(root, filepath.FromSlash(filepath.Clean(p))))
	}
}
