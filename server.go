package ddd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
)

// Server represents an HTTP server using Gorilla Mux
type Server struct {
	router     *mux.Router
	port       int
	host       string
	contexts   []*Context
	httpServer *http.Server
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
func (s *Server) WithContexts(contexts ...*Context) *Server {
	s.contexts = contexts
	return s
}

// Start initializes and starts the server
func (s *Server) Start() error {
	return s.StartWithContext(context.Background())
}

// StartWithContext starts the server with a context for graceful shutdown
func (s *Server) StartWithContext(ctx context.Context) error {
	// Register health check endpoint
	s.registerHealthCheckEndpoint()

	// Register context endpoints
	for _, ctxInstance := range s.contexts {
		s.registerContextEndpoints(ctxInstance)
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
		log.Printf("Starting server on %s", addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Wait for either context cancellation or server error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-ctx.Done():
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		return err
	}

	// Cleanup contexts
	for _, ctx := range s.contexts {
		if err := ctx.Destroy(); err != nil {
			log.Printf("Context cleanup error: %v", err)
		}
	}

	log.Println("Server shut down successfully")
	return nil
}

// registerHealthCheckEndpoint registers the health check endpoint
func (s *Server) registerHealthCheckEndpoint() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Status: UP")
	}).Methods("GET")
	log.Printf("Registered health check endpoint at /")
}

func (s *Server) registerContextEndpoints(ctx *Context) {
	contextName := ctx.Name()

	// Create a subrouter for the context
	contextRouter := s.router.PathPrefix("/" + contextName).Subrouter()

	// Get all endpoint types registered in the context
	s.registerEndpointTypes(ctx, contextRouter, contextName)
}

// registerEndpointTypes registers endpoint types that need to be resolved per request
func (s *Server) registerEndpointTypes(ctx *Context, router *mux.Router, contextName string) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	endpointType := reflect.TypeOf((*Endpoint)(nil)).Elem()

	for typ, implementations := range ctx.dependencies {
		// Check if the type implements the Endpoint interface
		if typ.Implements(endpointType) || reflect.PtrTo(typ).Implements(endpointType) {
			for name, info := range implementations {
				s.registerEndpointInstance(ctx, router, contextName, typ, name, info)
			}
		}
	}
}

// registerEndpointInstance registers a specific endpoint instance
func (s *Server) registerEndpointInstance(ctx *Context, router *mux.Router, contextName string, typ reflect.Type, name string, info *dependencyInfo) {
	// For request-scoped endpoints, we need to resolve them on each request
	if info.scope == Request {
		s.registerRequestScopedEndpoint(ctx, router, contextName, typ, name)
	} else {
		// For singleton/prototype, resolve once and register
		s.registerStaticEndpoint(ctx, router, contextName, typ, name)
	}
}

// registerRequestScopedEndpoint creates handlers that resolve the endpoint for each request
func (s *Server) registerRequestScopedEndpoint(ctx *Context, router *mux.Router, contextName string, typ reflect.Type, name string) {
	// We need to get the path from a temporary instance to register the route
	tempEndpoint, err := ctx.Resolve(typ, name)
	if err != nil {
		log.Printf("Failed to resolve temporary endpoint %s: %v", name, err)
		return
	}

	endpoint, ok := tempEndpoint.(Endpoint)
	if !ok {
		log.Printf("Resolved object is not an Endpoint: %s", name)
		return
	}

	path := endpoint.Path()

	// Discover available HTTP method handlers using reflection
	handlers := s.discoverHandlers(reflect.TypeOf(tempEndpoint))

	// Create wrapper handlers for each discovered method
	for method, methodName := range handlers {
		wrapperHandler := s.createRequestScopedHandler(ctx, typ, name, methodName, method)
		router.HandleFunc(path, wrapperHandler).Methods(string(method))
		log.Printf("Registered %s handler (request-scoped) for endpoint %s in context %s", method, path, contextName)
	}

	// Clean up the temporary instance
	ctx.ClearRequestScoped()
}

// registerStaticEndpoint registers singleton/prototype endpoints
func (s *Server) registerStaticEndpoint(ctx *Context, router *mux.Router, contextName string, typ reflect.Type, name string) {
	endpoint, err := ctx.Resolve(typ, name)
	if err != nil {
		log.Printf("Failed to resolve endpoint %s: %v", name, err)
		return
	}

	ep, ok := endpoint.(Endpoint)
	if !ok {
		log.Printf("Resolved object is not an Endpoint: %s", name)
		return
	}

	path := ep.Path()

	// Discover available HTTP method handlers using reflection
	handlers := s.discoverHandlers(reflect.TypeOf(endpoint))
	endpointValue := reflect.ValueOf(endpoint)

	// Register handlers for each discovered method
	for method, methodName := range handlers {
		handlerMethod := endpointValue.MethodByName(methodName)
		if !handlerMethod.IsValid() {
			continue
		}

		// Create a closure that captures the handler method
		handler := s.createStaticHandler(handlerMethod, methodName)
		router.HandleFunc(path, handler).Methods(string(method))
		log.Printf("Registered %s handler (static) for endpoint %s in context %s", method, path, contextName)
	}
}

// discoverHandlers uses reflection to find HTTP handler methods by convention
func (s *Server) discoverHandlers(typ reflect.Type) map[HttpMethod]string {
	handlers := make(map[HttpMethod]string)

	// Method name to HTTP method mapping
	methodMap := map[string]HttpMethod{
		"Get":     GET,
		"Post":    POST,
		"Put":     PUT,
		"Delete":  DELETE,
		"Patch":   PATCH,
		"Options": OPTIONS,
		"Head":    HEAD,
	}

	// Check each method on the type
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		methodName := method.Name

		// Check if this method name matches our HTTP method convention
		if httpMethod, exists := methodMap[methodName]; exists {
			// Verify the method signature: func(http.ResponseWriter, *http.Request)
			if s.isValidHandlerSignature(method.Type) {
				handlers[httpMethod] = methodName
			} else {
				log.Printf("Warning: Method %s has invalid handler signature, expected func(http.ResponseWriter, *http.Request)", methodName)
			}
		}
	}

	return handlers
}

// isValidHandlerSignature checks if a method has the correct HTTP handler signature
func (s *Server) isValidHandlerSignature(methodType reflect.Type) bool {
	// Should have 3 parameters: receiver, http.ResponseWriter, *http.Request
	if methodType.NumIn() != 3 {
		return false
	}

	// Should have no return values (or could be modified to allow error return)
	if methodType.NumOut() != 0 {
		return false
	}

	// Check parameter types
	responseWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	requestType := reflect.TypeOf((*http.Request)(nil))

	return methodType.In(1).Implements(responseWriterType) &&
		methodType.In(2) == requestType
}

// createRequestScopedHandler creates a handler that resolves the endpoint for each request
func (s *Server) createRequestScopedHandler(ctx *Context, typ reflect.Type, name string, methodName string, httpMethod HttpMethod) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Resolve the endpoint for this specific request (creates new instance)
		requestEndpoint, err := ctx.Resolve(typ, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve endpoint: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the handler method and call it
		endpointValue := reflect.ValueOf(requestEndpoint)
		handlerMethod := endpointValue.MethodByName(methodName)

		if !handlerMethod.IsValid() {
			http.Error(w, "Handler method not found", http.StatusInternalServerError)
			return
		}

		// Call the handler method
		handlerMethod.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		})

		// Clean up request-scoped dependencies after the request
		defer ctx.ClearRequestScoped()
	}
}

// createStaticHandler creates a handler for singleton/prototype endpoints
func (s *Server) createStaticHandler(handlerMethod reflect.Value, methodName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call the handler method directly
		handlerMethod.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		})
	}
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}
