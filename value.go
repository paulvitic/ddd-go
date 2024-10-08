package ddd

import (
	"fmt"
	"reflect"
)

type Value[T any] interface {
	Equals(other any) bool
	String() string
	Value() T
}

// ValueBase represents a value object with frozen properties.
type ValueBase[T any] struct {
	value T
}

// NewValue creates a new Value with the given value.
func NewValue[T any](value T) Value[T] {
	return &ValueBase[T]{
		value: value,
	}
}

// Equals checks if the value object is equal to another value object.
func (v *ValueBase[T]) Equals(other any) bool {
	if other == nil {
		return false
	}
	// check if the other is a value object
	_, ok := other.(Value[T])
	if ok {
		return deepEqual(v.value, other.(Value[T]).Value())
	}
	return false
}

// Value returns a read-only copy of the properties.
func (v *ValueBase[T]) Value() T {
	return v.value
}

// ToString returns the string value for the value object.
func (v *ValueBase[T]) String() string {
	return fmt.Sprint(v.value)
}

// deepEqual is a helper function for deep comparison. It can be implemented using a dedicated library or custom logic.
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
