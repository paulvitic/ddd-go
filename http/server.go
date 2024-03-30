package http

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Server interface {
	Start() error
	Stop() error
	RegisterEndpoint(endpoint Endpoint)
}

type server struct {
	srv    *http.Server
	router *mux.Router
}

func NewServer(addr string) Server {
	return &server{
		srv: &http.Server{
			Addr: addr,
		},
		router: mux.NewRouter(),
	}
}

func (s *server) Start() error {
	s.srv.Handler = s.router
	return s.srv.ListenAndServe()
}

func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set timeout for graceful shutdown
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *server) RegisterEndpoint(endpoint Endpoint) {
	s.router.HandleFunc(endpoint.Path(), endpoint.Handler()).Methods(endpoint.Methods()...)
}
