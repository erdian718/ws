package ws

import "strings"

// Router is the router.
type Router struct {
	key         string
	children    map[string]*Router
	middlewares []func(*Context) error
	handlers    map[string][]func(*Context) error
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
		r = &Router{
			children: make(map[string]*Router),
			handlers: make(map[string][](func(*Context) error)),
		}
		if len(key) > 1 && key[0] == ':' {
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

// func (a *Router) match(method string, path string, parameters map[string]string, handlers [][]func(*Context) error) (map[string]string, [][]func(*Context) error, error) {
// 	if len(a.middlewares) > 0 {
// 		handlers = append(handlers, a.middlewares)
// 	}

// 	if len(path) <= 0 || path[0] != '/' {
// 		return nil, nil, ErrNotFound
// 	}
// 	path = path[1:]
// 	i := strings.IndexRune(path, '/')

// 	var k string
// 	if i < 0 {
// 		k = path
// 	} else {
// 		k = path[:i]
// 	}

// 	if len(a.key) > 1 {
// 		if r, ok := a.children[a.key]; ok {
// 			parameters[a.key[1:]] = k
// 			if i < 0 {
// 				if hs, ok := r.handlers[method]; ok {
// 					return parameters, append(handlers, hs), nil
// 				}
// 				return nil, nil, ErrMethodNotAllowed
// 			}
// 			return r.match(method, path[i:], parameters, handlers)
// 		}
// 	}
// 	if r, ok := a.children[k]; ok {
// 		if i < 0 {
// 			if hs, ok := r.handlers[method]; ok {
// 				return parameters, append(handlers, hs), nil
// 			}
// 			return nil, nil, ErrMethodNotAllowed
// 		}
// 		return r.match(method, path[i:], parameters, handlers)
// 	}
// 	return nil, nil, ErrNotFound
// }
