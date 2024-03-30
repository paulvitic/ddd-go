package go_ddd

import (
	"github.com/stretchr/testify/assert" // Using testify for assertions
	"testing"
)

type testEntityWithStringId struct {
	Entity
	name string
}

func newTestEntityWithStringId(id string, name string) *testEntityWithStringId {
	return &testEntityWithStringId{
		NewEntity(NewID(id)),
		name,
	}
}

func (e *testEntityWithStringId) Name() string {
	return e.name
}

func (e *testEntityWithStringId) ChangeName(name string) {
	e.name = name
}

func TestNewEntity(t *testing.T) {
	testEntity := newTestEntityWithStringId("ID123", "value")
	assert.Equal(t, "ID123", testEntity.ID().String())
	assert.Equal(t, "value", testEntity.Name())
}

func TestEntity_UpdateName(t *testing.T) {
	testEntity := newTestEntityWithStringId("ID123", "value")
	testEntity.ChangeName("value2")
	assert.Equal(t, "value2", testEntity.Name())
}

func TestEntity_Equals(t *testing.T) {
	// Same object
	entity1 := newTestEntityWithStringId("ID123", "value")
	assert.True(t, entity1.Equals(*entity1))

	// Different object with same ID
	entity2 := newTestEntityWithStringId("ID123", "value2")
	assert.True(t, entity1.Equals(entity2))

	// Different ID same value
	entity3 := newTestEntityWithStringId("ID124", "value")
	assert.False(t, entity1.Equals(entity3))

	// Nil comparison
	assert.False(t, entity1.Equals(nil))

	// Different type
	assert.False(t, entity1.Equals(10)) // Integer comparison
}
