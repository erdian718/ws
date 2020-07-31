package ws

import (
	"log"
	"net/http"
	"strings"
)

// Router is the router.
type Router struct {
	key         string
	middlewares []func(*Context) error
	children    map[string]*Router
	handlers    map[string][]func(*Context) error
}

// New creates a new app.
func New() *Router {
	return &Router{
		children: make(map[string]*Router),
		handlers: make(map[string][]func(*Context) error),
	}
}

// Run runs the server at addr.
func (a *Router) Run(addr string) error {
	return http.ListenAndServe(addr, a)
}

// RunTLS runs the server at addr.
func (a *Router) RunTLS(addr string, certfile, keyfile string) error {
	return http.ListenAndServeTLS(addr, certfile, keyfile, a)
}

// Use uses the middlewares.
func (a *Router) Use(hs ...func(*Context) error) *Router {
	a.middlewares = append(a.middlewares, hs...)
	return a
}

// Handle registers the handler for the given pattern and method.
func (a *Router) Handle(method string, pattern string, hs ...func(*Context) error) *Router {
	r := a.Router(pattern)
	r.handlers[method] = append(r.handlers[method], hs...)
	return a
}

// Get registers the handler for the given pattern and method GET.
func (a *Router) Get(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle("GET", pattern, hs...)
}

// Post registers the handler for the given pattern and method POST.
func (a *Router) Post(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle("POST", pattern, hs...)
}

// Put registers the handler for the given pattern and method PUT.
func (a *Router) Put(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle("PUT", pattern, hs...)
}

// Patch registers the handler for the given pattern and method PATCH.
func (a *Router) Patch(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle("PATCH", pattern, hs...)
}

// Delete registers the handler for the given pattern and method DELETE.
func (a *Router) Delete(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle("DELETE", pattern, hs...)
}

// Router finds the router by pattern.
func (a *Router) Router(pattern string) *Router {
	if len(pattern) < 1 || pattern[0] != '/' {
		panic("ws: invalid router patttern: " + pattern)
	}
	pattern = pattern[1:]
	i := strings.IndexRune(pattern, '/')

	var key string
	if i < 0 {
		key = pattern
	} else {
		key = pattern[:i]
	}

	r, ok := a.children[key]
	if !ok {
		r = New()
		if len(key) > 0 && key[0] == ':' {
			if a.key != "" {
				panic("ws: conflict between parameters " + a.key + " and " + key)
			}
			a.key = key
		}
		a.children[key] = r
	}

	if i < 0 {
		return r
	}
	return r.Router(pattern[i:])
}

// Match matches the path.
func (a *Router) Match(path string) (*Router, string, string) {
	if len(path) < 1 || path[0] != '/' {
		return nil, "", ""
	}
	path = path[1:]
	i := strings.IndexRune(path, '/')

	var key, param string
	if i < 0 {
		key, path = path, ""
	} else {
		key, path = path[:i], path[i:]
	}
	if a.key != "" {
		key, param = a.key, key
	}
	return a.children[key], path, param
}

// ServeHTTP dispatches the request to the handler whose pattern matches the request URL.
func (a *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("ws:", e)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	path := r.URL.Path
	if len(path) < 1 || path[0] != '/' {
		path = "/" + path
	}

	err := (&Context{
		Request:        r,
		ResponseWriter: w,
		Path:           path,

		datas:  make(map[string]interface{}),
		params: make(map[string]string),
		router: a,
		index:  -len(a.middlewares),
	}).Next()

	if err != nil {
		if e, ok := err.(*StatusError); ok {
			w.WriteHeader(e.code)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
