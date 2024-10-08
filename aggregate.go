package ddd

import "reflect"

// Aggregate represents an aggregate root entity with event management functionalities
type Aggregate interface {
	Entity
	EventProducer
	AggregateType() string
}

type aggregate struct {
	Entity
	EventProducer
	aggType string
}

func (a *aggregate) AggregateType() string {
	return a.aggType
}

func (a *aggregate) ClearEvents() {
	a.EventProducer.ClearEvents()
}

// NewAggregate creates a new Aggregate instance
func NewAggregate[T any](id ID, aggregateType T) Aggregate {
	return &aggregate{
		Entity:        NewEntity(id),
		EventProducer: NewEventProducer(),
		aggType:       reflect.TypeOf(aggregateType).PkgPath() + "." + reflect.TypeOf(aggregateType).Name(),
	}
}
