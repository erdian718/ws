package ws

import (
	"fmt"
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

// Use uses the middlewares.
func (a *Router) Use(hs ...func(*Context) error) *Router {
	a.middlewares = append(a.middlewares, hs...)
	return a
}

// Handle registers the handler for the given pattern and method.
func (a *Router) Handle(method string, pattern string, hs ...func(*Context) error) *Router {
	r := a.Route(pattern)
	r.handlers[method] = append(r.handlers[method], hs...)
	return a
}

// Get registers the handler for the given pattern and method GET.
func (a *Router) Get(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle(http.MethodGet, pattern, hs...)
}

// Post registers the handler for the given pattern and method POST.
func (a *Router) Post(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle(http.MethodPost, pattern, hs...)
}

// Put registers the handler for the given pattern and method PUT.
func (a *Router) Put(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle(http.MethodPut, pattern, hs...)
}

// Patch registers the handler for the given pattern and method PATCH.
func (a *Router) Patch(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle(http.MethodPatch, pattern, hs...)
}

// Delete registers the handler for the given pattern and method DELETE.
func (a *Router) Delete(pattern string, hs ...func(*Context) error) *Router {
	return a.Handle(http.MethodDelete, pattern, hs...)
}

// Route finds the router by pattern.
func (a *Router) Route(pattern string) *Router {
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

	router, ok := a.children[key]
	if !ok {
		router = &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][]func(*Context) error),
		}
		if len(key) > 0 && key[0] == ':' {
			if a.key != "" {
				panic("ws: conflict between parameters " + a.key + " and " + key)
			}
			a.key = key
		}
		a.children[key] = router
	}

	if i < 0 {
		return router
	}
	return router.Route(pattern[i:])
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
	var err error
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("ws: %v", x)
		}
		finally(w, err)
	}()

	path := r.URL.Path
	if len(path) < 1 || path[0] != '/' {
		path = "/" + path
	}
	if r.Method == http.MethodOptions && path == "/*" {
		w.Header().Add("Allow", allAllow)
		err = Status(http.StatusOK, "")
		return
	}

	ctx := ctxPool.Get().(*Context)
	defer ctxPool.Put(ctx)
	ctx.Request = r
	ctx.ResponseWriter = w
	ctx.Path = path
	ctx.code = 0
	ctx.datas = make(map[string]interface{})
	ctx.params = make(map[string]string)
	ctx.querys = nil
	ctx.router = a
	ctx.index = -len(a.middlewares)
	err = ctx.Next()
}
