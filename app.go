package ws

import (
	"log"
	"net/http"
)

// App is the app entity.
type App struct {
	*Router
	maxMemory int64
}

// New creates a new app.
func New() *App {
	return &App{
		Router: &Router{
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

		app:    a,
		datas:  make(map[string]interface{}),
		params: make(map[string]string),
		router: a.Router,
		index:  -len(a.Router.middlewares),
	}).Next()

	if err != nil {
		var code int
		if e, ok := err.(*StatusError); ok {
			if e.isHTTPStatus {
				code = e.code
			} else if e.code == 0 {
				code = http.StatusBadRequest
			} else {
				code = http.StatusInternalServerError
			}
		} else {
			code = http.StatusInternalServerError
		}
		w.WriteHeader(code)
	}
}
