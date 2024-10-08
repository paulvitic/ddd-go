package ddd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
)

type Payload interface {
	Type() string
}

// HandlerFunc defines a function to execute the command executor.
// This function type is only used by `MiddlewareFunc`.
type HandlerFunc func(context.Context, Payload) (interface{}, error)

// MiddlewareFunc defines a function to process middleware.
// it receives the next Handler must return another executor
type MiddlewareFunc func(h HandlerFunc) HandlerFunc

// ServiceBus is the definition of how command should be handled
type ServiceBus interface {
	// Register assign a command to a command handle for future executions.
	Register(to string, handler HandlerFunc) error

	// Handlers returns all registered handlers.
	Handlers() map[string]HandlerFunc

	// Use adds middleware to the chain.
	Use(...MiddlewareFunc)

	// Dispatch send a given Command to its assigned command executor.
	Dispatch(context.Context, Payload) (interface{}, error)
}

type serviceBus struct {
	handlers   map[string]HandlerFunc
	middleware []MiddlewareFunc
}

// NewServiceBus creates a new service serviceBus.
func NewServiceBus() ServiceBus {
	return &serviceBus{
		handlers:   make(map[string]HandlerFunc),
		middleware: make([]MiddlewareFunc, 0),
	}
}

// Register assign a command to a command handle for future executions.
func (b *serviceBus) Register(to string, handler HandlerFunc) error {
	if _, ok := b.handlers[to]; !ok {
		b.handlers[to] = handler
		return nil
	}
	return errors.New("only one handler per message type is allowed")
}

// Handlers returns all registered handlers.
func (b *serviceBus) Handlers() map[string]HandlerFunc {
	return b.handlers
}

// Use adds middleware to the chain.
func (b *serviceBus) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

// Dispatch send a given Command to its assigned command executor.
func (b *serviceBus) Dispatch(ctx context.Context, msg Payload) (interface{}, error) {
	if err := validatePayload(msg); err != nil {
		return nil, err
	}

	if msgHandler, err := b.handler(msg.Type()); err != nil {
		return nil, err
	} else {
		h := applyMiddleware(msgHandler, b.middleware...)
		if res, err := h(ctx, msg); err != nil {
			return nil, err
		} else {
			return res, nil
		}
	}
}

func (b *serviceBus) handler(name string) (HandlerFunc, error) {
	if h, ok := b.handlers[name]; ok {
		return h, nil
	}

	return nil, errors.New(fmt.Sprintf("ServiceBus - handler not found for message %s", name))
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h
}

func validatePayload(msg Payload) error {
	value := reflect.ValueOf(msg)
	if value.Kind() != reflect.Ptr || !value.IsNil() && value.Elem().Kind() != reflect.Struct {
		return errors.New("command must be a struct ")
	}
	return nil
}

func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg Payload) (interface{}, error) {
			log.Printf("[ServiceBus] dispatching %s", msg.Type())
			return next(ctx, msg)
		}
	}
}
