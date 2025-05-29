package ddd

import (
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
	// Path returns the endpoint's URL path
	Path() string
}

func BindEndpoint(endpoint Endpoint, router *mux.Router) {
	path := endpoint.Path()
	handlers := requestHandlers(endpoint)

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

func requestHandlers(endpoint Endpoint) map[HttpMethod]string {
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
	for i := range typ.NumMethod() {
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

func GetContext(r *http.Request) *Context {
	if ctx, ok := r.Context().(*Context); !ok {
		panic("context nort found in request")
	} else {
		return ctx
	}
}
