package go_ddd

import (
	"github.com/google/uuid"
)

type ID interface {
	Equals(other any) bool
	String() string
}

type entityId[T any] struct {
	Value[T]
}

// NewID creates a new ID instance with the provided value.
func NewID(value interface{}) ID {
	if s, ok := value.(string); ok {
		return &entityId[string]{NewValue(s)}
	}
	if i, ok := value.(int); ok {
		return &entityId[int]{NewValue(i)}
	}
	panic("Invalid ID type")
}

func (e *entityId[T]) Equals(other any) bool {
	if other == nil {
		return false
	}
	_, ok := other.(ID)
	if ok {
		return e.Value.String() == other.(ID).String()
	}
	return false

}

func GenerateUUID() ID {
	return NewID(uuid.New().String())
}

type GenerateAutoIncrementId func(entityType string) ID
