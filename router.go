package ws

import (
	"strings"
)

// Router is the router.
type Router struct {
	param       string
	children    map[string]*Router
	middlewares []func(*Context) error
	handlers    map[string][]func(*Context) error
}

// Use uses the middlewares.
func (a *Router) Use(h func(*Context) error, hs ...func(*Context) error) *Router {
	middlewares := append(a.middlewares, h)
	a.middlewares = append(middlewares, hs...)
	return a
}

// Get registers the handler for the given pattern and method GET.
func (a *Router) Get(pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	return a.Handle("GET", pattern, h, hs...)
}

// Post registers the handler for the given pattern and method POST.
func (a *Router) Post(pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	return a.Handle("POST", pattern, h, hs...)
}

// Put registers the handler for the given pattern and method PUT.
func (a *Router) Put(pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	return a.Handle("PUT", pattern, h, hs...)
}

// Patch registers the handler for the given pattern and method PATCH.
func (a *Router) Patch(pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	return a.Handle("PATCH", pattern, h, hs...)
}

// Delete registers the handler for the given pattern and method DELETE.
func (a *Router) Delete(pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	return a.Handle("DELETE", pattern, h, hs...)
}

// Handle registers the handler for the given pattern and method.
func (a *Router) Handle(method string, pattern string, h func(*Context) error, hs ...func(*Context) error) *Router {
	r := a.Router(pattern)
	handlers := append(r.handlers[method], h)
	r.handlers[method] = append(handlers, hs...)
	return a
}

// Router finds the router by pattern.
func (a *Router) Router(pattern string) *Router {
	if len(pattern) <= 0 || pattern[0] != '/' {
		panic("ws: invalid router patttern: " + pattern)
	}
	pattern = pattern[1:]
	i := strings.IndexRune(pattern, '/')

	var k string
	if i < 0 {
		k = pattern
	} else {
		k = pattern[:i]
	}

	r, ok := a.children[k]
	if !ok {
		r = &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][](func(*Context) error)),
		}
		if len(k) > 1 && k[0] == ':' {
			if a.param != "" {
				panic("ws: conflict between parameters " + a.param + " and " + k)
			}
			a.param = k
		}
		a.children[k] = r
	}

	if i < 0 {
		return r
	}
	return r.Router(pattern[i:])
}

func (a *Router) match(method string, path string, parameters map[string]string, handlers [][]func(*Context) error) (map[string]string, [][]func(*Context) error, error) {
	if len(a.middlewares) > 0 {
		handlers = append(handlers, a.middlewares)
	}

	if len(path) <= 0 || path[0] != '/' {
		return nil, nil, ErrNotFound
	}
	path = path[1:]
	i := strings.IndexRune(path, '/')

	var k string
	if i < 0 {
		k = path
	} else {
		k = path[:i]
	}

	if len(a.param) > 1 {
		if r, ok := a.children[a.param]; ok {
			parameters[a.param[1:]] = k
			if i < 0 {
				if hs, ok := r.handlers[method]; ok {
					return parameters, append(handlers, hs), nil
				}
				return nil, nil, ErrMethodNotAllowed
			}
			return r.match(method, path[i:], parameters, handlers)
		}
	}
	if r, ok := a.children[k]; ok {
		if i < 0 {
			if hs, ok := r.handlers[method]; ok {
				return parameters, append(handlers, hs), nil
			}
			return nil, nil, ErrMethodNotAllowed
		}
		return r.match(method, path[i:], parameters, handlers)
	}
	return nil, nil, ErrNotFound
}
