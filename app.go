package ws

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// App is the app entity.
type App struct {
	router       *Router
	root         string
	dir          string
	maxMemory    int64
	errorHandler func(*http.Request, error)
}

// New creates a new app.
func New() *App {
	return &App{
		router: &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][]func(*Context) error),
		},
		maxMemory: 32 << 20,
		errorHandler: func(r *http.Request, err error) {
			log.Println(r.Method, r.URL, err)
		},
	}
}

// Static servers static file.
func (a *App) Static(root, dir string) *App {
	a.root = root
	a.dir = dir
	return a
}

// ErrorHandler sets error handler.
func (a *App) ErrorHandler(h func(*http.Request, error)) *App {
	a.errorHandler = h
	return a
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
			handleError(w, r, fmt.Errorf("%w: %v", ErrInternalServerError, e), a.errorHandler)
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

// Use uses the middlewares.
func (a *App) Use(h func(*Context) error, hs ...func(*Context) error) *App {
	middlewares := append(a.router.middlewares, h)
	a.router.middlewares = append(middlewares, hs...)
	return a
}

// Get registers the handler for the given pattern and method GET.
func (a *App) Get(pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Get(pattern, h, hs...)
	return a
}

// Post registers the handler for the given pattern and method POST.
func (a *App) Post(pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Post(pattern, h, hs...)
	return a
}

// Put registers the handler for the given pattern and method PUT.
func (a *App) Put(pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Put(pattern, h, hs...)
	return a
}

// Patch registers the handler for the given pattern and method PATCH.
func (a *App) Patch(pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Patch(pattern, h, hs...)
	return a
}

// Delete registers the handler for the given pattern and method DELETE.
func (a *App) Delete(pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Delete(pattern, h, hs...)
	return a
}

// Handle registers the handler for the given pattern and method.
func (a *App) Handle(method string, pattern string, h func(*Context) error, hs ...func(*Context) error) *App {
	a.router.Handle(method, pattern, h, hs...)
	return a
}

// Router finds the router by pattern.
func (a *App) Router(pattern string) *Router {
	if pattern == "" {
		return a.router
	}
	return a.router.Router(pattern)
}
