package ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProducer(t *testing.T) {
	ep := NewEventProducer()
	ep.RegisterEvent("aggType", NewID("ID123"), testEventPayload{Name: "value"})
	events := ep.Events()

	assert.Equal(t, 1, len(events))
	assert.Equal(t, "github.com/paulvitic/ddd-go.testEventPayload", events[0].Type())
}
