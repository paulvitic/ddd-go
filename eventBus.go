package ddd

import (
	"context"
	"log"
	"sync"
)

// EventBus dispatches events to registered handlers
type EventBus interface {
	// Subscribe registers a handler for an event type
	Subscribe(handlers []EventHandler)
	// Dispatch sends an event to all registered handlers
	Dispatch(ctx context.Context, event Event) error
}

// eventBus is the standard implementation of the EventBus interface
type eventBus struct {
	handlers      map[string][]HandleEvent
	handlersMutex sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() EventBus {
	return &eventBus{
		handlers: make(map[string][]HandleEvent),
	}
}

func (b *eventBus) Subscribe(handlers []EventHandler) {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	for _, handler := range handlers {
		subscriptions := handler.SubscribedTo()

		for eventType, handlerFunc := range subscriptions {
			if _, ok := b.handlers[eventType]; !ok {
				b.handlers[eventType] = []HandleEvent{}
			}

			b.handlers[eventType] = append(b.handlers[eventType], handlerFunc)
			log.Printf("Subscribed handler to %s event", eventType)
		}
	}
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
