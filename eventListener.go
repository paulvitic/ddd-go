package ddd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EventListener listens to events from the EventLog and dispatches them to the EventBus
type EventListener interface {
	// Start begins listening to the event log queue and dispatching events
	OnStart() error
	// Stop halts event listening and dispatching
	OnDestroy() error
	// IsRunning returns true if the listener is currently running
	IsRunning() bool
}

// eventListener is the standard implementation of the EventListener interface
type eventListener struct {
	log         *Logger
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
	WorkerCount int `json:"eventListenerWorkerCount"`
}

func NewEventListenerConfig() (*EventListenerConfig, error) {
	return Configuration[EventListenerConfig]("configs/properties.json")
}

// NewEventListener creates a new event listener
func NewEventListener(config *EventListenerConfig, eventLog EventLog, eventBus EventBus) EventListener {
	return &eventListener{
		log:         NewLogger(),
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

// OnStart begins listening to the event log queue and dispatching events
func (l *eventListener) OnStart() error {
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
	// Use background context for the workers since they manage their own lifecycle
	ctx := context.Background()
	for range l.workerCount {
		l.listenerWg.Add(1)
		go l.listen(ctx, *eventQueue)
	}

	l.log.Info("Event listener started with %d workers", l.workerCount)
	return nil
}

// OnDestroy halts event listening and dispatching
func (l *eventListener) OnDestroy() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return nil // Already stopped
	}

	close(l.stopCh)
	l.running = false

	// Wait for workers to finish with a reasonable timeout
	done := make(chan struct{})
	go func() {
		l.listenerWg.Wait()
		close(done)
	}()

	// Create a timeout context for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for workers to finish or timeout
	select {
	case <-done:
		l.log.Info("Event listener stopped")
		return nil
	case <-ctx.Done():
		l.log.Warn("Timeout waiting for event listener to stop")
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
			// Use a background context for the event dispatch
			dispatchCtx := context.Background()
			if err := l.eventBus.Dispatch(dispatchCtx, event); err != nil {
				l.log.Error("Error dispatching event %s: %v", event.Type(), err)
			}
		case <-l.stopCh:
			// Listener is being stopped
			return
		case <-ctx.Done():
			// Context was cancelled (shouldn't happen with background context, but good practice)
			return
		}
	}
}
