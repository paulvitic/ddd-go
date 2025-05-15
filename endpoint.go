package ddd

import (
	"net/http"
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
	// GetHandlers returns a map of HTTP methods to handler functions
	Handlers() map[HttpMethod]http.HandlerFunc
}
