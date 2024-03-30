package go_ddd

import "context"

type EventLog interface {
	EventsOf(aggregateID, aggregateType string) []Event
	Middleware() MiddlewareFunc
}

// InMemoryEventLog is an in-memory implementation of EventLog for testing purposes
type InMemoryEventLog struct {
	events map[string][]Event
}

//goland:noinspection GoUnusedExportedFunction
func NewInMemoryEventLog() *InMemoryEventLog {
	return &InMemoryEventLog{events: make(map[string][]Event)}
}

func (e *InMemoryEventLog) EventsOf(aggregateID, aggregateType string) []Event {
	eventKey := aggregateType + ":" + aggregateID
	events, ok := e.events[eventKey]
	if !ok {
		return []Event{}
	}
	return events
}

func (e *InMemoryEventLog) Middleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg Payload) (interface{}, error) {
			event := msg.(Event)
			e.append(event)
			return next(ctx, msg)
		}
	}
}

func (e *InMemoryEventLog) append(event Event) {
	eventKey := event.AggregateType() + ":" + event.AggregateID().String()
	events, ok := e.events[eventKey]
	if !ok {
		events = []Event{event}
	} else {
		events = append(events, event)
	}
	e.events[eventKey] = events
}
