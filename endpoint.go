package ddd

import (
	"net/http"
	"reflect"
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

func RequestHandlers(typ reflect.Type) map[HttpMethod]string {

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
