package ddd

import (
	"context"
	"errors"
	"fmt"
)

// EventLog is responsible for storing and providing access to events
// It acts as both the persistent store for events and the source of
// events for the event listener
type EventLog interface {
	// EventsOf retrieves all events for an aggregate
	EventsOf(aggregateID string, aggregateType string) []Event
	// Append adds a single event to the log
	Append(ctx context.Context, event Event) error
	// AppendFrom adds all events from an aggregate to the log
	AppendFrom(ctx context.Context, aggregate Aggregate) error
	// Queue returns the underlying event queue that can be consumed by an EventListener
	Queue() any
}

// InMemoryEventLogConfig contains configuration for in-memory event log
type InMemoryEventLogConfig struct {
	BufferSize int `json:"bufferSize"`
}

// inMemoryEventLog is an in-memory implementation of EventLog
// It persists events in memory and provides a channel for event consumption
type inMemoryEventLog struct {
	queue chan Event
	log   map[string][]Event
}

// NewInMemoryEventLog creates a new in-memory event log
func NewInMemoryEventLog(config InMemoryEventLogConfig) EventLog {
	return &inMemoryEventLog{
		queue: make(chan Event, config.BufferSize),
		log:   make(map[string][]Event),
	}
}

// EventsOf returns all events for an aggregate
func (e *inMemoryEventLog) EventsOf(aggregateID, aggregateType string) []Event {
	eventKey := aggregateType + ":" + aggregateID
	aggEvents, ok := e.log[eventKey]
	if !ok {
		return []Event{}
	}
	return aggEvents
}

// Append adds a single event to the log
func (e *inMemoryEventLog) Append(ctx context.Context, event Event) error {
	// Record the event in the log
	eventKey := event.AggregateType() + ":" + event.AggregateID().String()
	events, ok := e.log[eventKey]
	if !ok {
		events = []Event{event}
		e.log[eventKey] = events
	} else {
		e.log[eventKey] = append(events, event)
	}

	// Publish the event to the queue
	select {
	case e.queue <- event:
		// Event sent successfully
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, don't block but log warning
		return errors.New("event queue is full, event not published")
	}

	return nil
}

// AppendFrom adds all events from an aggregate to the log
func (e *inMemoryEventLog) AppendFrom(ctx context.Context, aggregate Aggregate) error {
	var errs []error
	for _, ev := range aggregate.GetAllEvents() {
		if err := e.Append(ctx, ev); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors encountered: %v", errs)
	}
	return nil
}

// Queue returns the underlying event queue
func (e *inMemoryEventLog) Queue() any {
	return &e.queue
}
