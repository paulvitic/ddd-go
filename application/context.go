package application

import (
	"fmt"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
)

type Context struct {
	name             string
	eventBus         ddd.EventBus
	commandBus       ddd.CommandBus
	queryBus         ddd.QueryBus
	eventPublisher   ddd.EventPublisher
	messageConsumers map[string]ddd.MessageConsumer
	queryEndpoints   map[string]*http.QueryEndpoint
	commandEndpoints map[string]*http.CommandEndpoint
}

func NewContext(options interface{}) *Context {
	queryBus := ddd.NewQueryBus()
	commandBus := ddd.NewCommandBus()
	eventBus := ddd.NewEventBus(commandBus)

	//var err error
	var name string
	var eventPublisher ddd.EventPublisher

	if options == nil {
		// if no option is provided, assign default name to context
		name = "default"
		// and use in-memory event publisher
		eventPublisher = inMemory.NewEventPublisher()

	} else if contextName, ok := options.(string); ok {
		// if argument is a string, assign it to context name
		name = contextName
		// and use in-memory event publisher
		eventPublisher = inMemory.NewEventPublisher()

	} else if config, ok := options.(Context); ok {
		if config.name != "" && config.name != "default" {
			name = config.name
		} else {
			name = "default"
		}
	}

	return &Context{
		name:             name,
		eventBus:         eventBus,
		commandBus:       commandBus,
		queryBus:         queryBus,
		eventPublisher:   eventPublisher,
		messageConsumers: make(map[string]ddd.MessageConsumer),
		queryEndpoints:   make(map[string]*http.QueryEndpoint),
		commandEndpoints: make(map[string]*http.CommandEndpoint),
	}
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
	if _, ok := c.queryEndpoints[endpoint.Path()]; ok {
		panic(fmt.Sprintf("Endpoint %s already registered", endpoint.Path()))
	}

	if endpoint.Methods() == nil {
		panic(fmt.Sprintf("Endpoint %s has no methods", endpoint.Path()))
	}

	endpoint.RegisterQueryBus(c.queryBus)
	c.queryEndpoints[endpoint.Path()] = endpoint
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
	if _, ok := c.commandEndpoints[endpoint.Path()]; ok {
		panic(fmt.Sprintf("Endpoint %s already registered", endpoint.Path()))
	}
	if endpoint.Methods() == nil {
		panic(fmt.Sprintf("Endpoint %s has no methods", endpoint.Path()))
	}

	endpoint.RegisterCommandBus(c.commandBus)
	c.commandEndpoints[endpoint.Path()] = endpoint
	return c
}

func (c *Context) RegisterMessageConsumer(consumer ddd.MessageConsumer) *Context {
	if _, ok := c.messageConsumers[consumer.Target()]; !ok {
		consumer.SetEventBus(c.eventBus)
		c.messageConsumers[consumer.Target()] = consumer
	}
	return c
}

func (c *Context) Start() error {
	//for _, consumer := range c.messageConsumers {
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
