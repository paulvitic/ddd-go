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
	// GetPath returns the endpoint's URL path
	Path() string
}

func RequestHandlers(endpoint Endpoint) map[HttpMethod]string {
	typ := reflect.TypeOf(endpoint)

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
			if isValidHandlerSignature(method.Type) {
				handlers[httpMethod] = methodName
			}
		}
	}

	return handlers
}

func BindEndpoint(
	endpoint Endpoint,
	router *mux.Router,
	name string,
	scope Scope,
	resolveFunc func(typ reflect.Type, options ...any) (any, error)) {

	path := endpoint.Path()
	handlers := RequestHandlers(endpoint)

	if scope == Request {
		for method, methodName := range handlers {
			// Capture variables for closure
			currentMethod := method
			currentMethodName := methodName

			wrapperHandler := func(w http.ResponseWriter, r *http.Request) {
				// defer c.ClearRequestScoped()

				// Resolve endpoint for this specific request
				requestEndpoint, err := resolveFunc(reflect.TypeOf(endpoint), name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to resolve endpoint: %v", err), http.StatusInternalServerError)
					return
				}
				callHandlerMethod(requestEndpoint, currentMethodName, w, r)
			}

			router.HandleFunc(path, wrapperHandler).Methods(string(currentMethod))
		}
	} else {
		for method, methodName := range handlers {
			// Capture variables for closure
			currentMethod := method
			currentMethodName := methodName

			handler := func(w http.ResponseWriter, r *http.Request) {
				callHandlerMethod(endpoint, currentMethodName, w, r)
			}

			router.HandleFunc(path, handler).Methods(string(currentMethod))
		}
	}
}

func callHandlerMethod(endpoint any, methodName string, w http.ResponseWriter, r *http.Request) {
	endpointValue := reflect.ValueOf(endpoint)
	handlerMethod := endpointValue.MethodByName(methodName)

	if !handlerMethod.IsValid() {
		http.Error(w, "Handler method not found", http.StatusInternalServerError)
		return
	}

	handlerMethod.Call([]reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(r),
	})
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
