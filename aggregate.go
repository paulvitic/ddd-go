package go_ddd

import "reflect"

// Aggregate represents an aggregate root entity with event management functionalities
type Aggregate interface {
	Entity
	EventProducer
	AggregateType() string
}

type aggregate struct {
	aggType string
	Entity
	EventProducer
}

func (a *aggregate) AggregateType() string {
	return a.aggType
}

// NewAggregate creates a new Aggregate instance
func NewAggregate[T any](id ID, aggregateType T) Aggregate {
	return &aggregate{
		reflect.TypeOf(aggregateType).PkgPath() + "." + reflect.TypeOf(aggregateType).Name(),
		NewEntity(id),
		NewEventProducer(),
	}
}
