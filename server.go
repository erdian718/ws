package ws

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Server is the server struct.
type Server struct {
	*Router
	server *http.Server
	closed chan struct{}
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

// Start starts the server at addr.
func (a *Server) Start() error {
	go a.start()
	err := a.server.ListenAndServe()
	<-a.closed
	return err
}

// StartTLS starts the server at addr.
func (a *Server) StartTLS(certfile, keyfile string) error {
	go a.start()
	err := a.server.ListenAndServeTLS(certfile, keyfile)
	<-a.closed
	return err
}

// Shutdown gracefully shuts down the server.
func (a *Server) Shutdown(ctx context.Context) error {
	log.Println("ws: shutdown server")
	return a.server.Shutdown(ctx)
}

func (a *Server) start() {
	addr := a.server.Addr
	if addr == "" {
		addr = "default addr"
	}
	log.Println("ws: start server at", addr)

	a.closed = make(chan struct{})
	defer close(a.closed)

	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt, os.Kill)
	<-sigint

	if err := a.Shutdown(context.Background()); err != nil {
		log.Println("ws:", err)
	}
}
