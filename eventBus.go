package ddd

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// EventBus dispatches events to registered handlers
type EventBus interface {
	// Subscribe registers a handler for an event type
	Subscribe(handlers []EventHandler)
	// Dispatch sends an event to all registered handlers
	Dispatch(cevent Event) error
}

// eventBus is the standard implementation of the EventBus interface
type eventBus struct {
	ctx *Context
	log
	handlers      map[string][]HandleEvent
	handlersMutex sync.RWMutex
	queue         chan Event
	running       bool
	stopCh        chan struct{}
	workerCount   int
	listenerWg    sync.WaitGroup
	mu            sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus(ctx *Context) EventBus {
	return &eventBus{
		handlers: make(map[string][]HandleEvent),
		queue:    make(chan Event, config.BufferSize),
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
func (b *eventBus) Dispatch(event Event) error {

	// Publish the event to the queue
	select {
	case b.queue <- event:
		// Event sent successfully
	// case <-b.ctx.Done():
	// 	return ctx.Err()
	default:
		// Queue is full, don't block but log warning
		return errors.New("event queue is full, event not published")
	}
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
		if err := handler(b.ctx, event); err != nil {
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

func (b *eventBus) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// OnStart begins listening to the event log queue and dispatching events
func (b *eventBus) OnStart() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil // Already running
	}

	b.running = true
	b.stopCh = make(chan struct{})

	// Start worker goroutines to process events
	// Use background context for the workers since they manage their own lifecycle
	ctx := context.Background()
	for range b.workerCount {
		b.listenerWg.Add(1)
		go b.listen(ctx, b.queue)
	}

	b.log.Info("Event listener started with %d workers", l.workerCount)
	return nil
}

// OnDestroy halts event listening and dispatching
func (b *eventBus) OnDestroy() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil // Already stopped
	}

	close(b.stopCh)
	b.running = false

	// Wait for workers to finish with a reasonable timeout
	done := make(chan struct{})
	go func() {
		b.listenerWg.Wait()
		close(done)
	}()

	// Create a timeout context for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for workers to finish or timeout
	select {
	case <-done:
		b.log.Info("Event listener stopped")
		return nil
	case <-ctx.Done():
		l.log.Warn("Timeout waiting for event listener to stop")
		return ctx.Err()
	}
}

// listen is the worker goroutine that listens to the event queue and dispatches events
func (b *eventBus) listen(ctx context.Context, eventQueue chan Event) {
	defer b.listenerWg.Done()

	for {
		select {
		case event, ok := <-eventQueue:
			if !ok {
				// Channel was closed
				return
			}
			// Dispatch the event to the event bus
			b.handlersMutex.RLock()
			handlers, ok := b.handlers[event.Type()]
			b.handlersMutex.RUnlock()

			if !ok {
				// No handlers for this event type
				return
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
		case <-b.stopCh:
			// Listener is being stopped
			return
		case <-ctx.Done():
			// Context was cancelled (shouldn't happen with background context, but good practice)
			return
		}
	}
}
