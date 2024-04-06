package ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockAggregate struct {
	Aggregate
	name       string
	valueProp  Value[struct{ name string }]
	entityProp *mockEntity
}

type mockEntity struct {
	Entity
	name string
}

func newMockEntity(id string, name string) *mockEntity {
	return &mockEntity{
		NewEntity(NewValue[string](id)),
		name,
	}
}

func newMockAggregate(name string, valueName string, entityName string) *mockAggregate {
	return &mockAggregate{
		Aggregate:  NewAggregate(GenerateUUID(), mockAggregate{}),
		name:       name,
		valueProp:  NewValue(struct{ name string }{valueName}),
		entityProp: newMockEntity("ID123", entityName),
	}
}

type nameUpdated struct {
	Name string
}

type valueReplaced struct {
	Name string
}

type entityNameUpdated struct {
	Name string
}

func (c *mockAggregate) Name() string {
	return c.name
}

func (c *mockAggregate) UpdateName(name string) {
	// Business logic
	c.name = name
	c.RegisterEvent(c.AggregateType(), c.ID(), nameUpdated{name})
}

func TestNewAggregate(t *testing.T) {
	agg := newMockAggregate("AAggregate", "AValue", "AEntity")
	expectedAggregateType := "github.com/paulvitic/ddd-go.mockAggregate"
	assert.Equal(t, expectedAggregateType, agg.AggregateType(),
		"Expected event.AggregateType to be %s, but got %s", expectedAggregateType, agg.AggregateType())
	assert.Equal(t, "AAggregate", agg.Name(), "Expected state.name to be 'AAggregate'")
}

func TestAggregate_UpdateName(t *testing.T) {
	agg := newMockAggregate("AAggregate", "AValue", "AEntity")
	agg.UpdateName("CompanyB")

	domainEvents := agg.Events()
	assert.Equal(t, 1, len(domainEvents),
		"Expected events count to be 1, but got %d", len(domainEvents))

	event := domainEvents[0]
	assert.Equal(t, agg.ID(), event.AggregateID(),
		"Expected event.AggregateID to be %s, but got %s", agg.ID(), event.AggregateID())

	assert.Equal(t, agg.AggregateType(), event.AggregateType(),
		"Expected event.AggregateType to be %s, but got %s", agg.AggregateType(), event.AggregateType())

	expectedEventType := "github.com/paulvitic/ddd-go.nameUpdated"
	assert.Equal(t, expectedEventType, event.Type(),
		"Expected event.Payload type to be %s, but got %s", expectedEventType, event.Type())

	//println(event.AggregateID())
	//println(event.ToJsonString())
}
