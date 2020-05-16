package ws

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// App is the app entity.
type App struct {
	root      string
	dir       string
	maxMemory int64
	router    *Router
}

// New creates a new app.
func New() *App {
	return &App{
		maxMemory: 32 << 20,
		router: &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][]func(*Context) error),
		},
	}
}

// Static servers static file.
func (a *App) Static(root, dir string) *App {
	a.root = root
	a.dir = dir
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
	path := r.URL.Path
	if path[0] != '/' {
		path = "/" + path
	}
	params, handlers, err := a.router.match(r.Method, path, make(map[string]string), nil)
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
	if errors.Is(err, ErrMethodNotAllowed) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if errors.Is(err, ErrNotFound) {
		path := r.URL.Path
		if len(path) <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if path[0] != '/' {
			path = "/" + path
		}
		if path[len(path)-1] == '/' {
			path = path + "index.html"
		}

		path = strings.TrimPrefix(filepath.Clean(filepath.FromSlash(path)), a.root)
		f, err := os.Open(filepath.Join(a.dir, path))
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, f.Name(), stat.ModTime(), f)
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
