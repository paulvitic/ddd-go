package ddd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

// EventListener listens to events from the EventLog and dispatches them to the EventBus
type EventListener interface {
	// Start begins listening to the event log queue and dispatching events
	Start(ctx context.Context) error
	// Stop halts event listening and dispatching
	Stop(ctx context.Context) error
	// IsRunning returns true if the listener is currently running
	IsRunning() bool
}

// eventListener is the standard implementation of the EventListener interface
type eventListener struct {
	eventLog    EventLog
	eventBus    EventBus
	running     bool
	stopCh      chan struct{}
	workerCount int
	listenerWg  sync.WaitGroup
	mu          sync.RWMutex
}

// EventListenerConfig contains configuration for the event listener
type EventListenerConfig struct {
	// WorkerCount is the number of workers that process events
	WorkerCount int `json:"workerCount"`
}

// NewEventListener creates a new event listener
func NewEventListener(config EventListenerConfig, eventLog EventLog, eventBus EventBus) EventListener {
	return &eventListener{
		eventLog:    eventLog,
		eventBus:    eventBus,
		stopCh:      make(chan struct{}),
		workerCount: config.WorkerCount,
	}
}

// IsRunning returns true if the listener is currently running
func (l *eventListener) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.running
}

// Start begins listening to the event log queue and dispatching events
func (l *eventListener) Start(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return nil // Already running
	}

	// Get the queue from the event log
	queue := l.eventLog.Queue()
	if queue == nil {
		return errors.New("event log queue is nil")
	}

	// Try to cast the queue to a channel of events
	eventQueue, ok := queue.(*chan Event)
	if !ok {
		return fmt.Errorf("event log queue is not a channel of events, got %T", queue)
	}

	l.running = true
	l.stopCh = make(chan struct{})

	// Start worker goroutines to process events
	for range l.workerCount {
		l.listenerWg.Add(1)
		go l.listen(ctx, *eventQueue)
	}

	log.Printf("Event listener started with %d workers", l.workerCount)
	return nil
}

// Stop halts event listening and dispatching
func (l *eventListener) Stop(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return nil // Already stopped
	}

	close(l.stopCh)
	l.running = false

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		l.listenerWg.Wait()
		close(done)
	}()

	// Wait for workers to finish or context to be cancelled
	select {
	case <-done:
		log.Println("Event listener stopped")
		return nil
	case <-ctx.Done():
		log.Println("Context cancelled while waiting for event listener to stop")
		return ctx.Err()
	}
}

// listen is the worker goroutine that listens to the event queue and dispatches events
func (l *eventListener) listen(ctx context.Context, eventQueue chan Event) {
	defer l.listenerWg.Done()

	for {
		select {
		case event, ok := <-eventQueue:
			if !ok {
				// Channel was closed
				return
			}
			// Dispatch the event to the event bus
			// Use a background context for the event dispatch to ensure it continues
			// even if the original context is cancelled
			dispatchCtx := context.Background()
			if err := l.eventBus.Dispatch(dispatchCtx, event); err != nil {
				log.Printf("Error dispatching event %s: %v", event.Type(), err)
			}
		case <-l.stopCh:
			// Listener is being stopped
			return
		case <-ctx.Done():
			// Context was cancelled
			return
		}
	}
}
