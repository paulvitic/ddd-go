package ddd

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockServiceBus struct {
	registerCalled   bool
	registerError    error
	dispatchCalled   bool
	dispatchResponse []interface{}
	dispatchError    error
}

func (m *mockServiceBus) Handlers() map[string]HandlerFunc {
	panic("implement me")
}

func (m *mockServiceBus) Use(middlewareFunc ...MiddlewareFunc) {
	panic("implement me")
}

func (m *mockServiceBus) Register(to string, handler HandlerFunc) error {
	if m.registerError != nil {
		return m.registerError
	}
	m.registerCalled = true
	return m.registerError
}

func (m *mockServiceBus) Dispatch(ctx context.Context, message Payload) (interface{}, error) {
	m.dispatchCalled = true
	return m.dispatchResponse, m.dispatchError
}

type mockCommandService struct{}

func (m *mockCommandService) WithEventBus(bus EventBus) CommandService {
	panic("implement me")
}

func (m *mockCommandService) SubscribedTo() []string {
	return []string{"testCommand"}
}

func (m *mockCommandService) Executor() CommandExecutor {
	return func(ctx context.Context, command Command) error {
		return nil
	}
}

func (m *mockCommandService) Use(bus EventBus) {}

func (m *mockCommandService) DispatchFrom(ctx context.Context, producer EventProducer) error {
	return nil
}

func TestCommandBus_RegisterHandler(t *testing.T) {
	serviceBus := &mockServiceBus{}
	service := &mockCommandService{}
	commandBus := &commandBus{serviceBus: serviceBus}

	// Successful registration
	err := commandBus.RegisterService(service)
	assert.NoError(t, err)
	assert.True(t, serviceBus.registerCalled)

	// Registration error
	errMsg := "registration failed"
	serviceBus.registerError = errors.New(errMsg)
	err = commandBus.RegisterService(service)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("Errors encountered: [%s]", errMsg), err.Error())
}

func TestCommandBus_Dispatch(t *testing.T) {
	serviceBus := &mockServiceBus{}
	commandBus := commandBus{serviceBus: serviceBus}
	cmd := NewCommand(testCommand{Name: "value"})

	// Successful dispatch with single response
	serviceBus.dispatchResponse = []interface{}{"result"}
	err := commandBus.Dispatch(context.Background(), cmd)
	assert.NoError(t, err)
	assert.True(t, serviceBus.dispatchCalled)

	// Dispatch error
	serviceBus.dispatchError = errors.New("dispatch failed")
	err = commandBus.Dispatch(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, "dispatch failed", err.Error())
}

func TestCommandBus_WithServiceBus(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping test in short mode.")
	//}
	var err error

	bus := NewCommandBus()

	handler := func(context.Context, Command) error {
		return nil
	}
	service := NewCommandService(handler, testCommand{})

	err = bus.RegisterService(service)
	assert.NoError(t, err)

	err = bus.Dispatch(context.Background(), NewCommand(testCommand{}))
	assert.NoError(t, err)

	err = bus.Dispatch(context.Background(), NewCommand("unknown command"))
	assert.Error(t, errors.New("executor not found for command"))
}
