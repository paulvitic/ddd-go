package http

import (
	"net/http"
)

type Endpoint interface {
	Path() string
	Methods() []string
	Handler() func(http.ResponseWriter, *http.Request)
}

type EndpointBase struct {
	path    string
	methods []string
}

func NewEndpoint(path string, methods []string) *EndpointBase {
	return &EndpointBase{
		path:    path,
		methods: methods,
	}
}

func (e *EndpointBase) Path() string {
	return e.path
}

func (e *EndpointBase) Methods() []string {
	return e.methods
}
