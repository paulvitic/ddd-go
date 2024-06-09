package context

import (
	"fmt"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
	"log"
)

func contextLog(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[info] Context: %s", msg), args...))
}

type Configuration struct {
	Name string `json:"name"`
}

type Context struct {
	Name             string
	profile          string
	eventBus         ddd.EventBus
	commandBus       ddd.CommandBus
	queryBus         ddd.QueryBus
	EventPublisher   ddd.EventPublisher
	MessageConsumers map[string]ddd.MessageConsumer
	QueryEndpoints   map[string]*http.QueryEndpoint
	CommandEndpoints map[string]*http.CommandEndpoint
}

func NewContext(config interface{}) *Context {
	queryBus := ddd.NewQueryBus()
	commandBus := ddd.NewCommandBus()
	eventBus := ddd.NewEventBus(commandBus)

	context := &Context{
		queryBus:         queryBus,
		commandBus:       commandBus,
		eventBus:         eventBus,
		MessageConsumers: make(map[string]ddd.MessageConsumer),
		QueryEndpoints:   make(map[string]*http.QueryEndpoint),
		CommandEndpoints: make(map[string]*http.CommandEndpoint),
	}

	if config != nil {
		if ctxOptions, ok := config.(Configuration); ok {
			context.configure(ctxOptions)
		} else {
			panic("Invalid config")
		}

	} else {
		context.configure(Configuration{
			Name: "default",
		})
	}

	return context
}

func (c *Context) configure(options Configuration) {
	if options.Name == "" {
		c.Name = "default"
	} else {
		c.Name = options.Name
	}

	c.EventPublisher = inMemory.NewEventPublisher()
}

func (c *Context) RegisterPolicy(policy ddd.Policy) *Context {
	if err := c.eventBus.RegisterPolicy(policy); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterView(view ddd.View) *Context {
	if err := c.eventBus.RegisterView(view); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterQueryService(service ddd.QueryService) *Context {
	if err := c.queryBus.RegisterService(service); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterQueryEndpoint(endpoint *http.QueryEndpoint) *Context {
	if _, ok := c.QueryEndpoints[endpoint.Path()]; ok {
		panic(fmt.Sprintf("Endpoint %s already registered", endpoint.Path()))
	}

	if endpoint.Methods() == nil {
		panic(fmt.Sprintf("Endpoint %s has no methods", endpoint.Path()))
	}

	endpoint.RegisterQueryBus(c.queryBus)
	c.QueryEndpoints[endpoint.Path()] = endpoint
	return c
}

func (c *Context) RegisterCommandService(service ddd.CommandService) *Context {
	service.WithEventBus(c.eventBus)
	if err := c.commandBus.RegisterService(service.WithEventBus(c.eventBus)); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterCommandEndpoint(endpoint *http.CommandEndpoint) *Context {
	if _, ok := c.CommandEndpoints[endpoint.Path()]; ok {
		panic(fmt.Sprintf("Endpoint %s already registered", endpoint.Path()))
	}
	if endpoint.Methods() == nil {
		panic(fmt.Sprintf("Endpoint %s has no methods", endpoint.Path()))
	}

	endpoint.RegisterCommandBus(c.commandBus)
	c.CommandEndpoints[endpoint.Path()] = endpoint
	return c
}

func (c *Context) RegisterMessageConsumer(consumer ddd.MessageConsumer) *Context {
	if _, ok := c.MessageConsumers[consumer.Target()]; !ok {
		consumer.SetEventBus(c.eventBus)
		c.MessageConsumers[consumer.Target()] = consumer
	}
	return c
}

func (c *Context) Start() error {
	//for _, consumer := range c.MessageConsumers {
	//	consumer.Start()
	//}
	return nil
}

func hasHttpMethod(source []string, target []string) bool {
	// Use a map for efficient lookups in the second array
	lookupMap := make(map[string]struct{})
	for _, val := range target {
		lookupMap[val] = struct{}{}
	}

	for _, val := range source {
		if _, ok := lookupMap[val]; ok {
			return true
		}
	}
	return false
}
