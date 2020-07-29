package ws

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// App is the app entity.
type App struct {
	*Router
	maxMemory int64
}

// New creates a new app.
func New() *App {
	return &App{
		router: &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][]func(*Context) error),
		},
		maxMemory: 32 << 20,
	}
}

// MaxMemory sets the max memory per request of the app.
func (a *App) MaxMemory(s int64) *App {
	a.maxMemory = s
	return a
}

// Run runs the app at addr.
func (a *App) Run(addr string) error {
	return http.ListenAndServe(addr, a)
}

// RunTLS runs the app at addr.
func (a *App) RunTLS(addr string, certfile, keyfile string) error {
	return http.ListenAndServeTLS(addr, certfile, keyfile, a)
}

// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			handleError(w, fmt.Errorf("%v", e))
		}
	}()

	path := r.URL.Path
	if len(path) <= 0 || path[0] != '/' {
		path = "/" + path
	}
	params, handlers, err := a.router.match(r.Method, path, make(map[string]string), nil)
	if err == nil {
		c := &Context{
			Request:        r,
			ResponseWriter: w,
			app:            a,
			data:           make(map[string]interface{}),
			params:         params,
			handlers:       handlers,
		}
		handleError(w, r, c.Next(), a.errorHandler)
		return
	}
	if errors.Is(err, ErrMethodNotAllowed) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if errors.Is(err, ErrNotFound) {
		if r.Method == "GET" && a.dir != "" {
			if path[len(path)-1] == '/' {
				path = path + "index.html"
			}
			path = filepath.Join(a.dir, filepath.Clean(filepath.FromSlash(strings.TrimPrefix(path, a.root))))
			err = serveFile(w, r, path)
		}
		handleError(w, r, err, a.errorHandler)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}
