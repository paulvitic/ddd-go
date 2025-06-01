package ddd

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

// HttpMethod represents HTTP methods as constants
type HttpMethod string

const (
	GET     HttpMethod = "GET"
	POST    HttpMethod = "POST"
	PUT     HttpMethod = "PUT"
	DELETE  HttpMethod = "DELETE"
	PATCH   HttpMethod = "PATCH"
	OPTIONS HttpMethod = "OPTIONS"
	HEAD    HttpMethod = "HEAD"
)

type Endpoint interface {
	OnInit() error
}

type endpoint struct {
	paths  []string
	value  any
	logger *Logger
	router *mux.Router
}

func NewEndpoint(value any, paths []string, logger *Logger, router *mux.Router) Endpoint {

	// Check if value is a pointer type
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("NewEndpoint: value must be a pointer type, got %T", value))
	}

	// Optional: Check if the pointer is not nil
	if rv.IsNil() {
		panic("NewEndpoint: value cannot be a nil pointer")
	}

	return &endpoint{
		paths:  paths,
		value:  value,
		logger: logger,
		router: router,
	}
}

func (e *endpoint) OnInit() error {
	for _, path := range e.paths {
		handlers := e.requestHandlers()

		for method, handler := range handlers {
			e.router.HandleFunc(path, handler).Methods(string(method))
			e.logger.Info("registered request handler at %s %s", string(method), path)
		}
	}
	return nil
}

// requestHandlers returns a map of HTTP methods to their actual handler functions
func (e *endpoint) requestHandlers() map[HttpMethod]http.HandlerFunc {
	typ := reflect.TypeOf(e.value)
	val := reflect.ValueOf(e.value)
	handlers := make(map[HttpMethod]http.HandlerFunc)

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
	for i := range typ.NumMethod() {
		method := typ.Method(i)
		methodName := method.Name

		// Check if this method name matches our HTTP method convention
		if httpMethod, exists := methodMap[methodName]; exists {
			// Verify the method signature: func(http.ResponseWriter, *http.Request)
			if isValidHandlerSignature(method.Type) {
				// Get the actual method and create a handler function
				methodValue := val.MethodByName(methodName)

				// Create the handler function that calls the method directly
				handler := func(w http.ResponseWriter, r *http.Request) {
					methodValue.Call([]reflect.Value{
						reflect.ValueOf(w),
						reflect.ValueOf(r),
					})
				}

				handlers[httpMethod] = handler
			}
		}
	}

	return handlers
}

// isValidHandlerSignature checks if a method has the correct HTTP handler signature
func isValidHandlerSignature(methodType reflect.Type) bool {
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

func GetContext(r *http.Request) *Context {
	if ctx := r.Context().Value(AppContextKey); ctx != nil {
		return ctx.(*Context)
	}
	return nil
}
