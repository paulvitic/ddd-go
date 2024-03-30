package go_ddd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestPolicy struct {
	Policy
}

// implement Test policy constructor
func NewTestPolicy() *TestPolicy {
	return &TestPolicy{
		NewPolicy([]string{"github.com/paulvitic/ddd-go.testEventPayload"}),
	}
}

func (p *TestPolicy) When(event Event) (Command, error) {
	switch event.Type() {
	case "github.com/paulvitic/ddd-go.testEventPayload":
		return NewCommand(testCommand{Name: "value"}), nil
	default:
		return nil, errors.New("unknown event type")
	}
}

func TestPolicy_subscription(t *testing.T) {
	policy := NewTestPolicy()
	event := NewEventProducer().
		RegisterEvent("aggType", NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()
	cmd, err := policy.When(event)
	assert.NoError(t, err)
	assert.Equal(t, "github.com/paulvitic/ddd-go.testCommand", cmd.Type())
}
