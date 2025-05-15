package ddd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Server represents an HTTP server using Gorilla Mux
type Server struct {
	router   *mux.Router
	port     int
	host     string
	contexts []*Context
}

// NewServer creates a new server instance
func NewServer(host string, port int) *Server {
	return &Server{
		router:   mux.NewRouter(),
		port:     port,
		host:     host,
		contexts: make([]*Context, 0),
	}
}

// WithContexts registers contexts with the server
func (s *Server) WithContexts(contexts []*Context) *Server {
	s.contexts = contexts
	return s
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Register health check endpoint
	s.registerHealthCheckEndpoint()

	// Register context endpoints
	for _, ctx := range s.contexts {
		s.registerContextEndpoints(ctx)
	}

	// Start the server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// registerHealthCheckEndpoint registers the health check endpoint
func (s *Server) registerHealthCheckEndpoint() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Status: UP")
	}).Methods("GET")
	log.Printf("Registered health check endpoint at /")
}

// registerContextEndpoints registers all endpoints for a context
func (s *Server) registerContextEndpoints(ctx *Context) {
	contextName := ctx.Name()
	endpoints := ctx.Endpoints()

	// Create a subrouter for the context
	contextRouter := s.router.PathPrefix("/" + contextName).Subrouter()

	for _, endpoint := range endpoints {
		path := endpoint.Path()
		handlers := endpoint.Handlers()

		// Register each handler with its specific HTTP method
		for method, handler := range handlers {
			contextRouter.HandleFunc(path, handler).Methods(string(method))
			log.Printf("Registered %s handler for endpoint %s in context %s", method, path, contextName)
		}
	}
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}
