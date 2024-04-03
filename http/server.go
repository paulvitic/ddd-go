package http

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server interface {
	Start()
	Stop() error
	RegisterEndpoint(endpoint Endpoint)
}

type server struct {
	srv    *http.Server
	router *mux.Router
}

func NewServer(addr string) Server {
	if addr == "" {
		addr = ":8080"
	}
	return &server{
		srv: &http.Server{
			Addr: addr,
		},
		router: mux.NewRouter(),
	}
}

func (s *server) Start() {
	logServerInfo("starting to listen at %v", s.srv.Addr)
	s.srv.Handler = s.router
	started := make(chan struct{})
	go func() {
		close(started)
		if err := s.srv.ListenAndServe(); err != nil {
			logServerWarning("%v", err)
		}
	}()
	<-started
	logServerInfo("started listening at %v", s.srv.Addr)
	return
}

func (s *server) Stop() error {
	logServerInfo("stopping listening at %v", s.srv.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set timeout for graceful shutdown
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *server) RegisterEndpoint(endpoint Endpoint) {
	path := "/api/" + endpoint.Path()
	logServerInfo("registering %v endpoints at %s", strings.Join(endpoint.Methods(), ", "), path)
	s.router.HandleFunc(path, endpoint.Handler()).Methods(endpoint.Methods()...)
}

func logServerInfo(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[info] HttpServer: %s", msg), args...))
}

func logServerWarning(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[warn] HttpServer: %s", msg), args...))
}
