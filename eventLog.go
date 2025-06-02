package ddd

import (
	"errors"
	"fmt"
	"sync"
)

// Enhanced EventLog interface with better error handling and querying
type EventLog interface {
	// EventsOf retrieves all events for an aggregate
	EventsOf(aggregateID string, aggregateType string) ([]Event, error)
	// EventsOfType retrieves all events of a specific type
	EventsOfType(eventType string) ([]Event, error)
	// Append adds a single event to the log
	Append(event Event) error
	// AppendFrom adds all events from an aggregate to the log
	AppendFrom(aggregate Aggregate) error
	// Close cleans up resources
	Close() error
}

// InMemoryEventLogConfig contains configuration for in-memory event log
type InMemoryEventLogConfig struct {
	BufferSize int `json:"inMemoryEventLogBufferSize"`
}

func NewInMemoryEventLogConfig() (*InMemoryEventLogConfig, error) {
	return Configuration[InMemoryEventLogConfig]("configs/properties.json")
}

// inMemoryEventLog is an enhanced in-memory implementation of EventLog
type inMemoryEventLog struct {
	logger *Logger
	// Aggregate-based storage: aggregateType:aggregateID -> []Event
	aggregateEvents map[string][]Event
	// Type-based storage: eventType -> []Event
	typeEvents map[string][]Event
	// Global event storage for complete event history
	allEvents []Event
	// Mutex for thread-safe operations
	mu sync.RWMutex
	// Optional: Configuration
	config *InMemoryEventLogConfig
}

// NewInMemoryEventLog creates a new in-memory event log
func NewInMemoryEventLog(config *InMemoryEventLogConfig) EventLog {
	if config == nil {
		config = &InMemoryEventLogConfig{BufferSize: 1000}
	}

	return &inMemoryEventLog{
		aggregateEvents: make(map[string][]Event),
		typeEvents:      make(map[string][]Event),
		allEvents:       make([]Event, 0, config.BufferSize),
		config:          config,
	}
}

// Middleware returns the middleware function for event logging
func (e *inMemoryEventLog) Middleware(next HandleEvent) HandleEvent {
	return func(event Event) error {
		// Log the event before passing to next middleware
		if err := e.Append(event); err != nil {
			e.logger.Error("Failed to append event to log: %v", err)
			// Decide whether to continue or fail the dispatch
			// For now, we'll continue but you might want to make this configurable
		}

		// Continue to next middleware/handler
		return next(event)
	}
}

// EventsOf returns all events for a specific aggregate
func (e *inMemoryEventLog) EventsOf(aggregateID, aggregateType string) ([]Event, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	eventKey := aggregateType + ":" + aggregateID
	events, ok := e.aggregateEvents[eventKey]
	if !ok {
		return []Event{}, nil
	}

	// Return a copy to prevent external modification
	result := make([]Event, len(events))
	copy(result, events)
	return result, nil
}

// EventsOfType returns all events of a specific type
func (e *inMemoryEventLog) EventsOfType(eventType string) ([]Event, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	events, ok := e.typeEvents[eventType]
	if !ok {
		return []Event{}, nil
	}

	// Return a copy to prevent external modification
	result := make([]Event, len(events))
	copy(result, events)
	return result, nil
}

// Append adds a single event to the log
func (e *inMemoryEventLog) Append(event Event) error {
	if event == nil {
		return errors.New("event cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Store by aggregate
	aggregateKey := event.AggregateType() + ":" + event.AggregateID().String()
	e.aggregateEvents[aggregateKey] = append(e.aggregateEvents[aggregateKey], event)

	// Store by event type
	eventType := event.Type()
	e.typeEvents[eventType] = append(e.typeEvents[eventType], event)

	// Store in global list
	e.allEvents = append(e.allEvents, event)

	return nil
}

// AppendFrom adds all events from an aggregate to the log
func (e *inMemoryEventLog) AppendFrom(aggregate Aggregate) error {
	if aggregate == nil {
		return errors.New("aggregate cannot be nil")
	}

	events := aggregate.GetAllEvents()
	if len(events) == 0 {
		return nil // No events to append
	}

	var errs []error
	for _, event := range events {
		if err := e.Append(event); err != nil {
			errs = append(errs, fmt.Errorf("failed to append event %s: %w", event.Type(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors encountered while appending events: %v", errs)
	}

	return nil
}

// Close cleans up resources (no-op for in-memory implementation)
func (e *inMemoryEventLog) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear all data
	e.aggregateEvents = make(map[string][]Event)
	e.typeEvents = make(map[string][]Event)
	e.allEvents = nil

	return nil
}

// GetAllEvents returns all events in the log (useful for debugging/testing)
func (e *inMemoryEventLog) GetAllEvents() []Event {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Event, len(e.allEvents))
	copy(result, e.allEvents)
	return result
}
