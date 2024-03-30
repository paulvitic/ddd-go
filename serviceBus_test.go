package go_ddd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert" // Using testify for assertions
)

type testMessage struct{}

func (t *testMessage) Type() string {
	return "test"
}

type mockService struct {
	subscribedTo []string
	handler      HandlerFunc
}

func (m *mockService) SubscribedTo() []string {
	return m.subscribedTo
}

func (m *mockService) Handler() HandlerFunc {
	return m.handler
}

func TestNewServiceBus(t *testing.T) {
	bus := NewServiceBus()
	assert.NotNil(t, bus)
	assert.Empty(t, bus.Handlers())
}

func TestServiceBus_Register(t *testing.T) {
	bus := NewServiceBus()
	assert.Empty(t, bus.Handlers())

	handler := func(context.Context, Payload) (interface{}, error) { return nil, nil }
	err := bus.Register("test1", handler)
	assert.NoError(t, err)
	err = bus.Register("test2", handler)
	assert.NoError(t, err)

	handlers := bus.Handlers()
	assert.NotEmpty(t, handlers)
	assert.Len(t, handlers, 2)
}

func TestServiceBus_InvalidRegistration(t *testing.T) {
	bus := NewServiceBus()
	handler := func(context.Context, Payload) (interface{}, error) { return nil, nil }
	err := bus.Register("test", handler)
	assert.NoError(t, err)

	err = bus.Register("test", handler)
	assert.Error(t, err)
}

func TestServiceBus_Dispatch(t *testing.T) {
	bus := NewServiceBus()

	calls := make([]string, 0)

	// Successful dispatch with multiple handlers and middleware
	handler := func(context.Context, Payload) (interface{}, error) {
		calls = append(calls, "handler")
		return "result", nil
	}

	err := bus.Register("test", handler)
	assert.NoError(t, err)

	res, err := bus.Dispatch(context.Background(), &testMessage{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"handler"}, calls)
	assert.Equal(t, "result", res)
}

func TestServiceBus_DispatchWithMiddleware(t *testing.T) {
	bus := NewServiceBus()

	calls := make([]string, 0)

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg Payload) (interface{}, error) {
			calls = append(calls, "middleware1")
			return next(ctx, msg)
		}
	}
	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg Payload) (interface{}, error) {
			calls = append(calls, "middleware2")
			return next(ctx, msg)
		}
	}
	bus.Use(Logger(), middleware1, middleware2)

	handler := func(context.Context, Payload) (interface{}, error) {
		calls = append(calls, "handler")
		return "result", nil
	}

	err := bus.Register("test", handler)
	assert.NoError(t, err)

	res, err := bus.Dispatch(context.Background(), &testMessage{})
	assert.NoError(t, err)
	assert.Equal(t, "result", res)
	assert.Equal(t, []string{"middleware1", "middleware2", "handler"}, calls)
}

func TestServiceBus_HandlerErrors(t *testing.T) {
	bus := NewServiceBus()

	handler := func(context.Context, Payload) (interface{}, error) { return nil, errors.New("executor error") }

	err := bus.Register("test", handler)
	assert.NoError(t, err)

	res, err := bus.Dispatch(context.Background(), &testMessage{})
	assert.Empty(t, res)
	assert.Error(t, err)
	assert.Equal(t, "executor error", err.Error())
}
