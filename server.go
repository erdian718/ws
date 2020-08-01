package ws

import (
	"context"
	"net/http"
	"time"
)

// Server is the server struct.
type Server struct {
	*Router
	server *http.Server
}

// New creates a new server.
func New() *Server {
	router := &Router{
		children: make(map[string]*Router),
		handlers: make(map[string][]func(*Context) error),
	}
	return &Server{
		Router: router,
		server: &http.Server{
			Handler: router,
		},
	}
}

// Server returns the http server.
func (a *Server) Server() *http.Server {
	return a.server
}

// Addr sets the addr.
func (a *Server) Addr(addr string) *Server {
	a.server.Addr = addr
	return a
}

// ReadHeaderTimeout sets the read header timeout.
func (a *Server) ReadHeaderTimeout(d time.Duration) *Server {
	a.server.ReadHeaderTimeout = d
	return a
}

// ReadTimeout sets the read timeout.
func (a *Server) ReadTimeout(d time.Duration) *Server {
	a.server.ReadTimeout = d
	return a
}

// WriteTimeout sets the write timeout.
func (a *Server) WriteTimeout(d time.Duration) *Server {
	a.server.WriteTimeout = d
	return a
}

// IdleTimeout sets the idle timeout.
func (a *Server) IdleTimeout(d time.Duration) *Server {
	a.server.IdleTimeout = d
	return a
}

// MaxHeaderBytes sets the max header bytes.
func (a *Server) MaxHeaderBytes(n int) *Server {
	a.server.MaxHeaderBytes = n
	return a
}

// Run runs the server at addr.
func (a *Server) Run(addr string) error {
	return a.server.ListenAndServe()
}

// RunTLS runs the server at addr.
func (a *Server) RunTLS(addr string, certfile, keyfile string) error {
	return a.server.ListenAndServeTLS(certfile, keyfile)
}

// Shutdown gracefully shuts down the app.
func (a *Server) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
