package ddd

import (
	"reflect"
	"sync"
	"time"
)

// Aggregate represents an aggregate root entity
// with event management functionalities
type Aggregate interface {
	Entity
	// Event management
	RaiseEvent(aggregateType string, aggregateID ID, payload any) Aggregate
	GetAllEvents() []Event
	GetFirstEvent() Event
	ClearEvents()
	AggregateType() string
}

type aggregate struct {
	Entity
	aggType string
	events  []Event
	mu      sync.Mutex
}

func (a *aggregate) AggregateType() string {
	return a.aggType
}

// RaiseEvent adds an event to the aggregate's event list
func (a *aggregate) RaiseEvent(aggregateType string, aggregateID ID, payload any) Aggregate {
	a.mu.Lock()
	defer a.mu.Unlock()

	eventType := EventType(payload)
	timeStamp := time.Now()
	a.events = append(
		a.events,
		&event{
			aggregateType,
			aggregateID,
			eventType,
			timeStamp,
			payload,
		})
	return a
}

// GetAllEvents returns all the events that have been raised and clears them.
func (a *aggregate) GetAllEvents() []Event {
	e := a.events
	a.mu.Lock()
	defer a.mu.Unlock()

	a.events = make([]Event, 0)
	return e
}

// GetFirstEvent returns and removes the first event from the aggregate's event list
func (a *aggregate) GetFirstEvent() Event {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.events) == 0 {
		return nil
	}

	pop := a.events[0]
	a.events = a.events[1:]
	return pop
}

// ClearEvents removes all events from the aggregate
func (a *aggregate) ClearEvents() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.events = make([]Event, 0)
}

// NewAggregate creates a new Aggregate instance
func NewAggregate[T any](id ID, aggregateType T) Aggregate {
	return &aggregate{
		Entity:  NewEntity(id),
		aggType: reflect.TypeOf(aggregateType).PkgPath() + "." + reflect.TypeOf(aggregateType).Name(),
		events:  make([]Event, 0),
	}
}
