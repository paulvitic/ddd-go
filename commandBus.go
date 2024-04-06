package ddd

import (
	"context"
	"errors"
	"fmt"
)

type CommandBus interface {
	RegisterService(service CommandService) error
	Dispatch(ctx context.Context, command Command) error
	Use(middleware MiddlewareFunc)
}

type commandBus struct {
	serviceBus ServiceBus
}

func NewCommandBus() CommandBus {
	return &commandBus{
		NewServiceBus(),
	}
}

func (c *commandBus) RegisterService(service CommandService) error {
	var errs []error
	for _, s := range service.SubscribedTo() {
		handlerFunc := func(ctx context.Context, command Payload) (interface{}, error) {
			err := service.Executor()(ctx, command.(Command))
			return err == nil, err
		}
		err := c.serviceBus.Register(s, handlerFunc)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}

func (c *commandBus) Dispatch(ctx context.Context, command Command) error {
	_, err := c.serviceBus.Dispatch(ctx, command)
	if err != nil {
		return err
	}
	return nil
}

func (c *commandBus) Use(middleware MiddlewareFunc) {
	c.serviceBus.Use(middleware)
}
