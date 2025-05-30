package ddd

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// Middleware represents a middleware function that can wrap the dispatch process
type Middleware func(next HandleEvent) HandleEvent

// EventBus dispatches events to registered handlers with middleware support
type EventBus interface {
	// Subscribe registers a handler for an event type
	Subscribe(handlers []EventHandler)
	// Dispatch sends an event to all registered handlers through the middleware pipeline
	Dispatch(event Event) error
	// WithMiddleware adds middleware to the dispatch pipeline
	WithMiddleware(middleware ...Middleware) EventBus
	// Start begins the event bus operations
	Start() error
	// Stop gracefully shuts down the event bus
	Stop() error
	// IsRunning returns whether the event bus is currently running
	IsRunning() bool
}

// eventBus is the standard implementation of the EventBus interface
type eventBus struct {
	ctx           *Context
	logger        *Logger
	handlers      map[string][]HandleEvent
	handlersMutex sync.RWMutex
	queue         chan Event
	running       bool
	stopCh        chan struct{}
	workerCount   int
	listenerWg    sync.WaitGroup
	mu            sync.RWMutex
	middleware    []Middleware
	dispatchChain HandleEvent
}

// NewEventBus creates a new event bus
func NewEventBus(ctx *Context) EventBus {
	ctx.logger.Info("Creating new EventBus")

	eb := &eventBus{
		ctx:         ctx,
		logger:      ctx.logger,
		handlers:    make(map[string][]HandleEvent),
		queue:       make(chan Event, 100), // Buffer size may come from configuration
		workerCount: 1,                     // Default to 1 worker, could be configurable
		middleware:  make([]Middleware, 0),
	}

	// Initialize the dispatch chain with the core dispatch logic
	eb.buildDispatchChain()

	return eb
}

// WithMiddleware adds middleware to the dispatch pipeline
func (b *eventBus) WithMiddleware(middleware ...Middleware) EventBus {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Append new middleware to existing middleware
	b.middleware = append(b.middleware, middleware...)

	// Rebuild the dispatch chain with new middleware
	b.buildDispatchChain()

	return b
}

// buildDispatchChain constructs the middleware chain with the core dispatch logic at the end
func (b *eventBus) buildDispatchChain() {
	// Core dispatch function (the final handler in the chain)
	coreDispatch := func(ctx *Context, event Event) error {
		return b.coreDispatch(ctx, event)
	}

	// Build the chain by wrapping from right to left (last middleware wraps first)
	b.dispatchChain = coreDispatch
	for i := len(b.middleware) - 1; i >= 0; i-- {
		b.dispatchChain = b.middleware[i](b.dispatchChain)
	}
}

// coreDispatch is the original dispatch logic, now called at the end of the middleware chain
func (b *eventBus) coreDispatch(ctx context.Context, event Event) error {
	// Add event to queue for async processing
	select {
	case b.queue <- event:
		// Event queued successfully
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full
		return errors.New("event queue is full, event not published")
	}
}

// Subscribe registers handlers for event types
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

// Dispatch sends an event through the middleware pipeline
func (b *eventBus) Dispatch(event Event) error {
	// Execute the middleware chain
	return b.dispatchChain(b.ctx, event)
}

// Start begins the event bus operations
func (b *eventBus) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil // Already running
	}

	b.running = true
	b.stopCh = make(chan struct{})

	// Start worker goroutines to process events from the queue
	ctx := context.Background()
	for range b.workerCount {
		b.listenerWg.Add(1)
		go b.listen(ctx, b.queue)
	}

	b.logger.Info("Event bus started with %d workers", b.workerCount)
	return nil
}

// Stop gracefully shuts down the event bus
func (b *eventBus) Stop() error {
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
		b.logger.Info("Event bus stopped")
		return nil
	case <-ctx.Done():
		b.logger.Warn("Timeout waiting for event bus to stop")
		return ctx.Err()
	}
}

// IsRunning returns whether the event bus is currently running
func (b *eventBus) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// listen is the worker goroutine that processes events from the queue
func (b *eventBus) listen(ctx context.Context, eventQueue chan Event) {
	defer b.listenerWg.Done()

	for {
		select {
		case event, ok := <-eventQueue:
			if !ok {
				// Channel was closed
				return
			}
			// Process the event by calling registered handlers
			b.processEvent(event)

		case <-b.stopCh:
			// Event bus is being stopped
			return

		case <-ctx.Done():
			// Context was cancelled
			return
		}
	}
}

// processEvent executes all registered handlers for an event
func (b *eventBus) processEvent(event Event) {
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
		if err := handler(b.ctx, event); err != nil {
			log.Printf("Error handling event %s: %v", event.Type(), err)
			errs = append(errs, err)
			// Continue processing other handlers even if one fails
		}
	}

	if len(errs) > 0 {
		log.Printf("Errors encountered while processing event %s: %d errors", event.Type(), len(errs))
	}
}
