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

// MaxMemory sets the max memory per request of the app.
func (a *App) MaxMemory(s int64) *App {
	a.maxMemory = s
	return a
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
		if err := c.Next(); err != nil {
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
		}
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

// Run runs the app at addr.
func (a *App) Run(addr string) error {
	return http.ListenAndServe(addr, a)
}

// RunTLS runs the app at addr.
func (a *App) RunTLS(addr string, certfile, keyfile string) error {
	return http.ListenAndServeTLS(addr, certfile, keyfile, a)
}
