package application

import (
	"errors"
	"fmt"
	ddd "github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/amqp"
	"github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	contexts   map[string]*Context
	httpServer http.Server
}

func NewServer(option interface{}) *Server {
	httpServer := http.NewServer(":8080")

	if option == nil {
		logServerInfo("starting with default configuration")
	} else {
		profile, ok := option.(string)
		if ok {
			logServerInfo("starting with profile: " + profile + " from configuration file")
		}
		config, ok := option.(Configuration)
		if ok {
			logServerInfo("starting with port " + config.Port)
		}
	}

	return &Server{
		contexts:   make(map[string]*Context),
		httpServer: httpServer,
	}
}

func (a *Server) WithContext(context *Context) *Server {
	logServerInfo("registering %s context", context.name)
	if a.contexts[context.name] != nil {
		logServerWarning("context %s already exists", context.name)
		return a
	}
	a.contexts[context.name] = context
	return a
}

func (a *Server) WithHttpServer(server *http.Server) *Server {
	a.httpServer = *server
	return a
}

func (a *Server) Start() error {
	a.registerHttpEndpoints()
	if err := a.startMessageConsumers(); err != nil {
		return err
	}

	a.httpServer.Start()

	sigs := make(chan os.Signal, 1)
	// Register the channel to receive interrupt signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Wait for a signal
	<-sigs

	fmt.Println("Exiting application")
	return nil
}

func (a *Server) registerHttpEndpoints() {
	logServerInfo("registering http endpoints")
	for _, context := range a.contexts {
		for _, endpoint := range context.queryEndpoints {
			a.httpServer.RegisterEndpoint(endpoint)
		}
		for _, endpoint := range context.commandEndpoints {
			a.httpServer.RegisterEndpoint(endpoint)
		}
	}
}

func (a *Server) startMessageConsumers() error {
	for _, context := range a.contexts {
		for _, consumer := range context.messageConsumers {
			if err := a.startMessageConsumer(context, consumer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Server) startMessageConsumer(context *Context, consumer ddd.MessageConsumer) {
	logServerInfo("starting %s context %s message consumer", context.name, consumer.Target())
	target := consumer.Target()
	if a.contexts[target] != nil {
		targetQueue := a.contexts[target].eventPublisher.Queue()
		if channelQueue, ok := targetQueue.(*chan string); ok {
			consumer = inMemory.MessageConsumer(consumer, channelQueue)
		} else if amqpConfig, ok := targetQueue.(amqp.Configuration); ok {
			consumer = amqp.MessageConsumer(consumer, amqpConfig)
		} else {
			return errors.New(fmt.Sprintf("target context %s message queue type for %s context message consumer is not recognized", target, context.name))
		}
	}

	if err := consumer.Start(); err != nil {
		return err
	}
	return nil
}

func logServerInfo(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[info] AppServer: %s", msg), args...))
}

func logServerWarning(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[warn] AppServer: %s", msg), args...))
}
