package ddd

import (
	"context"
	"fmt"
	"os"

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
	// Add context and cancel function for proper shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

type ServerConfig struct {
	// WorkerCount is the number of workers that process events
	Host string `json:"serverHost"`
	Port int    `json:"serverPort"`
}

func NewServerConfig(configPath ...string) *ServerConfig {
	var path string
	if configPath == nil {
		path = os.Getenv("DDD_SERVER_CONFIG_PATH")
		if path == "" {
			path = "configs/properties.json"
		}
	} else {
		path = configPath[0]
	}
	config, err := Configuration[ServerConfig](path)
	if err != nil {
		panic(err)
	}
	return config
}

// NewServer creates a new server instance
func NewServer(serverConfig *ServerConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		logger:   NewLogger(),
		port:     serverConfig.Port,
		host:     serverConfig.Host,
		contexts: make([]*Context, 0),
		router:   mux.NewRouter(),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// WithContexts registers contexts with the server
func (s *Server) WithContexts(contextFacories ...ContextFactory) *Server {
	for _, contextFactory := range contextFacories {
		context := contextFactory(s.ctx, s.router)
		s.contexts = append(s.contexts, context)
	}
	return s
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Register health check endpoint
	s.registerHealthCheck()

	// Start all contexts
	for _, ctx := range s.contexts {
		ctx.Start()
	}

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

	// Wait for either context cancellation or server error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-s.ctx.Done():
		return s.shutdownHTTPServer()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.cancel != nil {
		// Cancel the context to signal Start() to exit
		s.cancel()
	}
	return nil
}

// shutdownHTTPServer handles the actual HTTP server shutdown
func (s *Server) shutdownHTTPServer() error {
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

	s.logger.Info("registered health check endpoint at GET /")
}
