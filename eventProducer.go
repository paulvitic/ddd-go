package go_ddd

import (
	"sync"
	"time"
)

type EventProducer interface {
	RegisterEvent(aggregateType string, aggregateID ID, payload any) EventProducer
	Events() []Event
	GetFirst() Event
}

type eventProducer struct {
	mu     sync.Mutex
	events []Event
}

func NewEventProducer() EventProducer {
	return &eventProducer{
		events: make([]Event, 0),
	}
}

// Events returns all the events that have been registered. It also clears the events.
func (ep *eventProducer) Events() []Event {
	e := ep.events
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.events = make([]Event, 0)
	return e
}

func (ep *eventProducer) RegisterEvent(aggregateType string, aggregateID ID, payload any) EventProducer {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	eventType := EventType(payload)
	timeStamp := time.Now()
	ep.events = append(
		ep.events,
		&event{aggregateType,
			aggregateID,
			eventType,
			timeStamp,
			payload})
	return ep
}

func (ep *eventProducer) GetFirst() Event {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	pop := ep.events[0]
	ep.events = ep.events[1:]
	return pop
}
