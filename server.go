package ddd

import (
	"context"
	"fmt"

	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Server represents an HTTP server using Gorilla Mux
type Server struct {
	logger     *Logger
	port       int
	host       string
	contexts   []*Context
	router     *mux.Router
	httpServer *http.Server
}

// NewServer creates a new server instance
// TODO: Use a Configuration with host port and active contexts
func NewServer(host string, port int) *Server {
	return &Server{
		logger:   NewLogger(),
		port:     port,
		host:     host,
		contexts: make([]*Context, 0),
		router:   mux.NewRouter(),
	}
}

// WithContexts registers contexts with the server
func (s *Server) WithContexts(contexts ...*Context) *Server {
	s.contexts = contexts
	return s
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}

// Start initializes and starts the server
func (s *Server) Start() error {

	s.bindEndpoins()
	// Register health check endpoint
	s.registerHealthCheck()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		s.logger.Info("Starting server on %s", addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	serverCtx := context.Background()
	// Wait for either context cancellation or server error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-serverCtx.Done():
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error: %v", err)
		return err
	}

	// Cleanup contexts
	for _, ctx := range s.contexts {
		if err := ctx.Destroy(); err != nil {
			s.logger.Error("Context cleanup error: %v", err)
		}
	}

	s.logger.Info("Server shut down successfully")
	return nil
}

// registerHealthCheck registers the health check endpoint
func (s *Server) registerHealthCheck() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Status: UP")
	}).Methods("GET")

	s.logger.Info("Registered health check endpoint at /")
}

func (s *Server) bindEndpoins() {
	for _, context := range s.contexts {
		context.BindEndpoints(s.router)
	}
}
