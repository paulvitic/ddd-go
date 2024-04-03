package application

import (
	"fmt"
	"github.com/paulvitic/ddd-go"
	dddhttp "github.com/paulvitic/ddd-go/http"
	"github.com/paulvitic/ddd-go/inMemory"
	"net/http"
)

type Context struct {
	name             string
	eventBus         go_ddd.EventBus
	commandBus       go_ddd.CommandBus
	queryBus         go_ddd.QueryBus
	eventPublisher   go_ddd.EventPublisher
	messageConsumers map[string]go_ddd.MessageConsumer
	endpoints        map[string]dddhttp.Endpoint
}

func NewContext(options interface{}) *Context {
	commandBus := go_ddd.NewCommandBus()
	queryBus := go_ddd.NewQueryBus()
	eventBus := go_ddd.NewEventBus(commandBus)

	//var err error
	var name string
	var eventPublisher go_ddd.EventPublisher

	if options == nil {
		name = "default"
		eventPublisher = inMemory.NewEventPublisher()

	} else if contextName, ok := options.(string); ok {
		name = contextName
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
		messageConsumers: make(map[string]go_ddd.MessageConsumer),
		endpoints:        make(map[string]dddhttp.Endpoint),
	}
}

func (c *Context) EventBus() go_ddd.EventBus {
	return c.eventBus
}

func (c *Context) RegisterQueryService(service go_ddd.QueryService) *Context {
	if err := c.queryBus.RegisterService(service); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterCommandService(service go_ddd.CommandService) *Context {
	service.WithEventBus(c.eventBus)
	if err := c.commandBus.RegisterService(service.WithEventBus(c.eventBus)); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterPolicy(policy go_ddd.Policy) *Context {
	if err := c.eventBus.RegisterPolicy(policy); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterView(view go_ddd.View) *Context {
	if err := c.eventBus.RegisterView(view); err != nil {
		panic(err)
	}
	return c
}

func (c *Context) RegisterMessageConsumer(consumer go_ddd.MessageConsumer) *Context {
	if _, ok := c.messageConsumers[consumer.Target()]; !ok {
		consumer.SetEventBus(c.eventBus)
		c.messageConsumers[consumer.Target()] = consumer
	}
	return c
}

func (c *Context) RegisterHttpEndpoint(endpoint dddhttp.Endpoint) *Context {
	if _, ok := c.endpoints[endpoint.Path()]; ok {
		panic(fmt.Sprintf("Endpoint %s already registered", endpoint.Path()))
	}
	if endpoint.Methods() == nil {
		panic(fmt.Sprintf("Endpoint %s has no methods", endpoint.Path()))
	}
	if hasHttpMethod(endpoint.Methods(), []string{http.MethodGet}) {
		endpoint.RegisterQueryBus(c.queryBus)
	}
	if hasHttpMethod(endpoint.Methods(), []string{http.MethodPost, http.MethodPut, http.MethodDelete}) {
		endpoint.RegisterCommandBus(c.commandBus)
	}
	c.endpoints[endpoint.Path()] = endpoint
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
