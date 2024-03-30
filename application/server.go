package application

import (
	"errors"
	"fmt"
	"github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
)

type Server struct {
	contexts   map[string]*Context
	httpServer http.Server
}

func NewServer(option interface{}) (*Server, error) {
	httpServer := http.NewServer(":8080")

	if option == nil {

	} else {
		profile, ok := option.(string)
		if ok {
			print("Starting Server with profile: " + profile + " from configuration file")
		}
		config, ok := option.(Configuration)
		if ok {
			print("Starting Server with configuration: " + config.Port)
		}
	}

	return &Server{
		contexts:   make(map[string]*Context),
		httpServer: httpServer,
	}, nil
}

func (a *Server) WithContext(context *Context) *Server {
	a.contexts[context.name] = context
	return a
}

func (a *Server) WithHttpServer(server *http.Server) *Server {
	return a
}

func (a *Server) Start() error {
	if err := a.registerHttpEndpoints(); err != nil {
		return err
	}
	for current, context := range a.contexts {
		for target, consumer := range context.messageConsumers {
			if a.contexts[target] != nil {
				targetQueue := a.contexts[target].eventPublisher.Queue()
				if channelQueue, ok := targetQueue.(*chan string); ok {
					context.messageConsumers[target] = inMemory.MessageConsumer(context.messageConsumers[target], channelQueue)
				} else {
					return errors.New(fmt.Sprintf("target context %s message queue type for %s context message consumer is not recognized", target, current))
				}
			}
			if err := consumer.Start(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Server) registerHttpEndpoints() error {
	for _, context := range a.contexts {
		for _, endpoint := range context.endpoints {
			a.httpServer.RegisterEndpoint(endpoint)
		}
	}
	return nil
}
