package inMemory

import (
	"context"
	"github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testEventPayload struct {
	Name string
}

func TestInMemoryEventPublisher(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping test in short mode.")
	//}
	publisher := NewEventPublisher()
	t.Cleanup(func() {
		publisher.Close()
	})

	event := go_ddd.NewEventProducer().
		RegisterEvent("aggType", go_ddd.NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()

	_, err := publisher.Middleware()(func(ctx context.Context, msg go_ddd.Payload) (interface{}, error) {
		return nil, nil
	})(context.Background(), event)
	assert.NoError(t, err)

	expected, err := event.ToJsonString()
	assert.NoError(t, err)

	actual := <-*publisher.Queue().(*chan string)
	assert.Equal(t, expected, actual)
}
