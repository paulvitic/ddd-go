package server

import (
	"errors"
	"fmt"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/amqp"
	"github.com/paulvitic/ddd-go/context"
	"github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Configuration struct {
	Host           string   `json:"host,omitempty"`
	Port           int      `json:"port"`
	ActiveContexts []string `json:"activeContexts"`
}

type Server struct {
	config     *Configuration
	contexts   map[string]*context.Context
	httpServer http.Server
}

func NewServer(config *Configuration) *Server {
	defaultPort := 8080
	if config.Port == 0 {
		config.Port = defaultPort
	}
	serverAddress := fmt.Sprintf("%s:%d", config.Host, config.Port)
	httpServer := http.NewServer(serverAddress)

	return &Server{
		config:     config,
		contexts:   make(map[string]*context.Context),
		httpServer: httpServer,
	}
}

func (a *Server) WithContext(context *context.Context) *Server {
	logServerInfo("registering %s context", context.Name)
	if a.contexts[context.Name] != nil {
		logServerWarning("context %s already exists", context.Name)
		return a
	}
	a.contexts[context.Name] = context
	return a
}

func (a *Server) WithHttpServer(server *http.Server) *Server {
	a.httpServer = *server
	return a
}

func (a *Server) Start() error {
	a.registerHttpEndpoints()
	a.startMessageConsumers()
	a.httpServer.Start()

	sigs := make(chan os.Signal, 1)
	// Register the channel to receive interrupt signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Wait for a signal
	<-sigs

	fmt.Println("Exiting server")
	return nil
}

func (a *Server) registerHttpEndpoints() {
	logServerInfo("registering http endpoints")
	for _, ctx := range a.contexts {
		for _, endpoint := range ctx.QueryEndpoints {
			a.httpServer.RegisterEndpoint(endpoint)
		}
		for _, endpoint := range ctx.CommandEndpoints {
			a.httpServer.RegisterEndpoint(endpoint)
		}
	}
}

func (a *Server) startMessageConsumers() {
	for _, ctx := range a.contexts {
		for _, consumer := range ctx.MessageConsumers {
			if err := a.startMessageConsumer(ctx, consumer); err != nil {
				panic(err)
			}
		}
	}
}

func (a *Server) startMessageConsumer(context *context.Context, consumer ddd.MessageConsumer) error {
	logServerInfo("starting %s context %s message consumer", context.Name, consumer.Target())
	target := consumer.Target()
	if a.contexts[target] != nil {
		targetQueue := a.contexts[target].EventPublisher.Queue()
		if channelQueue, ok := targetQueue.(*chan string); ok {
			consumer = inMemory.MessageConsumer(consumer, channelQueue)
		} else if amqpConfig, ok := targetQueue.(amqp.Configuration); ok {
			consumer = amqp.MessageConsumer(consumer, amqpConfig)
		} else {
			return errors.New(fmt.Sprintf(
				"target context %s message queue type for %s context message consumer is not recognized",
				target, context.Name))
		}
	}

	if err := consumer.Start(); err != nil {
		return err
	}
	return nil
}

func logServerInfo(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf("[info] AppServer: %s", msg), args...)
}

func logServerWarning(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf("[warn] AppServer: %s", msg), args...)
}
