package ddd

import (
	"context"
	"log"
	"sync"
)

// EventHandler defines a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

// EventBus dispatches events to registered handlers
type EventBus interface {
	// Subscribe registers a handler for an event type
	Subscribe(eventType string, handler EventHandler) EventBus
	// Dispatch sends an event to all registered handlers
	Dispatch(ctx context.Context, event Event) error
}

// eventBus is the standard implementation of the EventBus interface
type eventBus struct {
	handlers      map[string][]EventHandler
	handlersMutex sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() EventBus {
	return &eventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for an event type
func (b *eventBus) Subscribe(eventType string, handler EventHandler) EventBus {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	if _, ok := b.handlers[eventType]; !ok {
		b.handlers[eventType] = []EventHandler{}
	}

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	log.Printf("Subscribed handler to %s event", eventType)
	return b
}

// Dispatch sends an event to all registered handlers
func (b *eventBus) Dispatch(ctx context.Context, event Event) error {
	b.handlersMutex.RLock()
	handlers, ok := b.handlers[event.Type()]
	b.handlersMutex.RUnlock()

	if !ok {
		// No handlers for this event type
		return nil
	}

	// Execute all handlers for this event
	var errs []error

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("Error handling event %s: %v", event.Type(), err)
			errs = append(errs, err)
			// We continue processing other handlers even if one fails
		}
	}

	if len(errs) > 0 {
		log.Printf("Errors encountered while dispatching event %s: %d errors", event.Type(), len(errs))
	}

	return nil
}
