package ddd

import (
	"context"
	"errors"
	"fmt"
)

type QueryBus interface {
	RegisterService(service QueryService) error
	Dispatch(ctx context.Context, query Query) (QueryResponse, error)
	Use(middleware MiddlewareFunc)
}

type queryBus struct {
	serviceBus ServiceBus
}

func NewQueryBus() QueryBus {
	return &queryBus{
		serviceBus: NewServiceBus(),
	}
}

func (c *queryBus) RegisterService(service QueryService) error {
	var errs []error
	for _, s := range service.SubscribedTo() {
		handlerFunc := func(ctx context.Context, msg Payload) (interface{}, error) {
			return service.Executor()(ctx, msg.(Query))
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

func (c *queryBus) Dispatch(ctx context.Context, query Query) (QueryResponse, error) {
	res, err := c.serviceBus.Dispatch(ctx, query)
	if err != nil {
		return nil, err
	}

	qr, ok := res.(QueryResponse)
	if ok {
		return qr, nil
	} else {
		return nil, errors.New("QueryBus.Dispatch() expects a QueryResponse")
	}
}

func (c *queryBus) Use(middleware MiddlewareFunc) {
	c.serviceBus.Use(middleware)
}
