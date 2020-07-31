package logger

import (
	"log"

	"github.com/ofunc/ws"
)

// New creates a logger middleware.
func New(root string) func(*ws.Context) error {
	return func(ctx *ws.Context) error {
		err := ctx.Next()
		if err != nil {
			if _, ok := err.(*ws.StatusError); !ok {
				log.Println(err)
			}
		}
		return err
	}
}
