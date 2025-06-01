package ddd

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// Middleware represents a middleware function that can wrap the dispatch process
type Middleware func(next HandleEvent) HandleEvent

// eventBus is the standard implementation of the EventBus interface
type EventBus struct {
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
func NewEventBus(ctx *Context) *EventBus {
	eb := &EventBus{
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

func (b *EventBus) Init() {
	// TODO Find and add event bus middleware from the context
	handlers, err := ResolveAll[EventHandler](b.ctx)
	if err != nil {
		panic("can not get event handlers")
	}
	b.Subscribe(handlers)
}

// WithMiddleware adds middleware to the dispatch pipeline
func (b *EventBus) WithMiddleware(middleware ...Middleware) *EventBus {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Append new middleware to existing middleware
	b.middleware = append(b.middleware, middleware...)

	// Rebuild the dispatch chain with new middleware
	b.buildDispatchChain()

	return b
}

// buildDispatchChain constructs the middleware chain with the core dispatch logic at the end
func (b *EventBus) buildDispatchChain() {
	// Core dispatch function (the final handler in the chain)
	coreDispatch := func(event Event) error {
		return b.coreDispatch(event)
	}

	// Build the chain by wrapping from right to left (last middleware wraps first)
	b.dispatchChain = coreDispatch
	for i := len(b.middleware) - 1; i >= 0; i-- {
		b.dispatchChain = b.middleware[i](b.dispatchChain)
	}
}

// coreDispatch is the original dispatch logic, now called at the end of the middleware chain
func (b *EventBus) coreDispatch(event Event) error {
	// Check if event bus is running
	b.mu.RLock()
	running := b.running
	b.mu.RUnlock()

	if !running {
		return errors.New("event bus is not running")
	}

	// Add event to queue for async processing
	select {
	case b.queue <- event:
		// Event queued successfully
		return nil
	default:
		// Queue is full - this is non-blocking
		return errors.New("event queue is full, event not published")
	}
}

// Subscribe registers handlers for event types
func (b *EventBus) Subscribe(handlers []EventHandler) {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	for _, handler := range handlers {
		subscriptions := handler.SubscribedTo()

		for eventType, handlerFunc := range subscriptions {
			if _, ok := b.handlers[eventType]; !ok {
				b.handlers[eventType] = []HandleEvent{}
			}

			b.handlers[eventType] = append(b.handlers[eventType], handlerFunc)

			eventTypeParts := strings.Split(eventType, ".")
			b.logger.Info("subscribed handler to event %s", eventTypeParts[len(eventTypeParts)-1])
		}
	}
}

// Dispatch sends an event through the middleware pipeline
func (b *EventBus) Dispatch(event Event) error {
	// Execute the middleware chain
	return b.dispatchChain(event)
}

// Start begins the event bus operations
func (b *EventBus) Start() error {
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
func (b *EventBus) Stop() error {
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
func (b *EventBus) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// listen is the worker goroutine that processes events from the queue
func (b *EventBus) listen(ctx context.Context, eventQueue chan Event) {
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
func (b *EventBus) processEvent(event Event) {
	b.handlersMutex.RLock()
	handlers, ok := b.handlers[event.Type()]
	b.handlersMutex.RUnlock()

	if !ok {
		// No handlers for this event type
		return
	}

	// Execute all handlers for this event
	var errs []error
	for _, handle := range handlers {
		if err := handle(event); err != nil {
			b.logger.Error("Error handling event %s: %v", event.Type(), err)
			errs = append(errs, err)
			// Continue processing other handlers even if one fails
		}
	}

	if len(errs) > 0 {
		b.logger.Error("Errors encountered while processing event %s: %d errors", event.Type(), len(errs))
	}
}
