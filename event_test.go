package go_ddd

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testEventPayload struct {
	Name string
}

func TestDomainEvent(t *testing.T) {
	event := NewEventProducer().
		RegisterEvent("aggType", NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()
	str, err := event.ToJsonString()
	assert.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(str), &data)
	assert.NoError(t, err)

	assert.Equal(t, "aggType", data["aggregate_type"])
	assert.Equal(t, "ID123", data["aggregate_id"])
	assert.Equal(t, "github.com/paulvitic/ddd-go.testEventPayload", data["event_type"])
}

func TestEventFromJsonString(t *testing.T) {
	event1 := NewEventProducer().
		RegisterEvent("aggType", NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()
	str, err := event1.ToJsonString()
	assert.NoError(t, err)

	event2, err := EventFromJsonString(str)
	assert.NoError(t, err)

	assert.Equal(t, event1.AggregateType(), event2.AggregateType())
	assert.Equal(t, event1.AggregateID(), event2.AggregateID())
	assert.Equal(t, event1.Type(), event2.Type())

	payload := MapEventPayload(event2, testEventPayload{})
	assert.Equal(t, "value", payload.Name)
}
