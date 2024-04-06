package ddd

import (
	"context"
	"errors"
	"fmt"
)

type EventBus interface {
	Handler() HandlerFunc
	Dispatch(ctx context.Context, event Event) error
	DispatchFrom(ctx context.Context, producer EventProducer) error
	Use(middleware MiddlewareFunc)
	RegisterView(view View) error
	RegisterPolicy(policy Policy) error
}

type eventBus struct {
	serviceBus ServiceBus
	commandBus CommandBus
	views      map[string][]View

	policies map[string][]Policy
}

func NewEventBus(commandBus CommandBus) EventBus {
	return &eventBus{
		NewServiceBus(),
		commandBus,
		make(map[string][]View),
		make(map[string][]Policy),
	}
}

func (c *eventBus) Handler() HandlerFunc {
	return func(ctx context.Context, msg Payload) (interface{}, error) {
		event, ok := msg.(Event)
		if ok {
			var errs []error
			if err := c.dispatchToViews(event); err != nil {
				errs = append(errs, err)
			}
			if err := c.dispatchToPolicies(event); err != nil {
				errs = append(errs, err)
			}
			if len(errs) > 0 {
				return nil, errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
			}
			return true, nil
		} else {
			return ok, errors.New("eventBus.Handler() expects an Event")
		}
	}
}

func (c *eventBus) Dispatch(ctx context.Context, event Event) error {
	_, err := c.serviceBus.Dispatch(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

func (c *eventBus) DispatchFrom(ctx context.Context, producer EventProducer) error {
	var errs []error
	for _, ev := range producer.Events() {
		if err := c.Dispatch(ctx, ev); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}

func (c *eventBus) Use(middleware MiddlewareFunc) {
	c.serviceBus.Use(middleware)
}

func (c *eventBus) RegisterView(view View) error {
	var errs []error
	for _, subscribedTo := range view.SubscribedTo() {
		c.views[subscribedTo] = append(c.views[subscribedTo], view)
		err := c.serviceBus.Register(subscribedTo, c.Handler())
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}

func (c *eventBus) RegisterPolicy(policy Policy) error {
	var errs []error
	for _, subscribedTo := range policy.SubscribedTo() {
		c.policies[subscribedTo] = append(c.policies[subscribedTo], policy)
		err := c.serviceBus.Register(subscribedTo, c.Handler())
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}

func (c *eventBus) dispatchToViews(event Event) error {
	var errs []error
	views, ok := c.views[event.Type()]
	if ok {
		for _, view := range views {
			if err := view.MutateWhen(event); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}

func (c *eventBus) dispatchToPolicies(event Event) error {
	var errs []error
	policies, ok := c.policies[event.Type()]
	if ok {
		for _, policy := range policies {
			cmd, err := policy.When(event)
			if err != nil {
				errs = append(errs, err)
			}
			if cmd != nil {
				if err := c.commandBus.Dispatch(context.Background(), cmd); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if len(errs) > 0 {
		return errors.New("Errors encountered: " + fmt.Sprintf("%v", errs))
	}
	return nil
}
