package ws

import (
	"errors"
	"net/http"
)

// App is the app entity.
type App struct {
	maxMemory int64
	Router    *Router
}

// New creates a new app.
func New() *App {
	return &App{
		maxMemory: 32 << 20,
		Router: &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][]func(*Context) error),
		},
	}
}

// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, handlers, err := a.Router.match(r.Method, r.URL.Path, make(map[string]string), nil)
	if err == nil {
		c := &Context{
			Request:        r,
			ResponseWriter: w,
			Data:           make(map[string]interface{}),
			app:            a,
			params:         params,
			handlers:       handlers,
		}
		c.Next()
		return
	}
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, ErrMethodNotAllowed) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}
