package inMemory

import (
	"context"
	"errors"
	ddd "github.com/paulvitic/ddd-go"
)

// EventLog implementation
type EventLog struct {
	events     map[string]map[string][]ddd.Event
	middleware []ddd.MiddlewareFunc
}

// NewInMemoryEventLog creates a new instance
func NewInMemoryEventLog() *EventLog {
	return &EventLog{
		events: make(map[string]map[string][]ddd.Event),
	}
}

// EventsOf retrieves events for an aggregate ID and type
func (l *EventLog) EventsOf(aggregateID, aggregateType string) []ddd.Event {
	aggregateEvents, ok := l.events[aggregateType]
	if !ok {
		return nil
	}
	return aggregateEvents[aggregateID]
}

// Middleware returns the accumulated middleware for handling events
func (l *EventLog) Middleware() ddd.MiddlewareFunc {
	return func(h ddd.HandlerFunc) ddd.HandlerFunc {
		return func(ctx context.Context, payload ddd.Payload) (interface{}, error) {
			event, ok := payload.(ddd.Event)
			if ok {
				l.Append(event)
				return true, nil
			}
			return nil, errors.New("payload is not an Event")
		}
	}
}

// Append appends an event to the log
func (l *EventLog) Append(event ddd.Event) {
	aggregateEvents, ok := l.events[event.AggregateType()]
	if !ok {
		aggregateEvents = make(map[string][]ddd.Event)
		l.events[event.AggregateType()] = aggregateEvents
	}
	aggregateEvents[event.AggregateID().String()] = append(aggregateEvents[event.AggregateID().String()], event)
}
