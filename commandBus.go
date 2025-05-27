package ddd

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
)

// CommandBus is responsible for routing commands to their appropriate handlers
type CommandBus interface {
	// Subscribe registers command handlers with the bus
	Subscribe(handlers []CommandHandler) error
	// Dispatch sends a command to its registered handler
	Dispatch(ctx context.Context, command Command) error
}

// CommandHandler defines a handler that can process specific command types
type CommandHandler interface {
	// SubscribedTo returns a map of command types to handler functions
	SubscribedTo() map[string]func(context.Context, Command) error
}

// commandBus is the default implementation of CommandBus
type commandBus struct {
	handlers map[string]func(context.Context, Command) error
	logger   *log.Logger
	mutex    sync.RWMutex
}

// NewCommandBus creates a new command bus instance
func NewCommandBus() CommandBus {
	return &commandBus{
		handlers: make(map[string]func(context.Context, Command) error),
		logger:   log.New(log.Writer(), "CommandBus: ", log.LstdFlags),
	}
}

// Subscribe registers all command handlers with the bus
func (c *commandBus) Subscribe(handlers []CommandHandler) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var errs []error

	for _, handler := range handlers {
		subscriptions := handler.SubscribedTo()

		for cmdType, handlerFunc := range subscriptions {
			// Check if a handler already exists
			if _, exists := c.handlers[cmdType]; exists {
				err := fmt.Errorf("handler already registered for command type: %s", cmdType)
				errs = append(errs, err)
				c.logger.Printf("Warning: %v", err)
				continue
			}

			c.handlers[cmdType] = handlerFunc
			c.logger.Printf("Subscribed %s to %s command",
				reflect.TypeOf(handler).Elem().Name(),
				cmdType)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors registering command handlers: %v", errs)
	}

	return nil
}

// Dispatch sends a command to its registered handler
func (c *commandBus) Dispatch(ctx context.Context, command Command) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if command == nil {
		return fmt.Errorf("cannot dispatch nil command")
	}

	cmdType := command.Type()
	c.logger.Printf("Dispatching %s", cmdType)

	// Debug log the command details
	if debugEnabled {
		c.logger.Printf("Debug - %s: %+v", cmdType, command.Body())
	}

	// Get the handler for this command type
	handler, exists := c.handlers[cmdType]
	if !exists {
		return fmt.Errorf("no handler registered for command type: %s", cmdType)
	}

	// Execute the handler with context
	return handler(ctx, command)
}

// Debug flag to enable/disable detailed logging
var debugEnabled = false

// SetDebug enables or disables detailed debug logging
func SetDebug(enabled bool) {
	debugEnabled = enabled
}
