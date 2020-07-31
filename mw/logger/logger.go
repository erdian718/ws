package logger

import (
	"log"
	"net/http"

	"github.com/ofunc/ws"
)

// New creates a logger middleware.
func New() func(*ws.Context) error {
	return func(ctx *ws.Context) error {
		err := ctx.Next()
		if err != nil {
			if e, ok := err.(*ws.StatusError); ok {
				if e.Code() != http.StatusInternalServerError {
					return err
				}
			}
			log.Println(err)
		}
		return err
	}
}
