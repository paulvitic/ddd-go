package inMemory

import (
	"context"
	"errors"
	"github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testCommand struct {
	Name string
}

// MockEventBus is a mock implementation of the EventBus interface for testing purposes.
type MockEventBus struct {
	t             *testing.T
	dispatchCalls []*funcCall // Store expected Dispatch calls
}

type funcCall struct {
	aggregateID   string
	aggregateType string
	eventType     string
}

// NewMockEventBus creates a new MockEventBus instance.
func NewMockEventBus(t *testing.T) *MockEventBus {
	return &MockEventBus{
		t:             t,
		dispatchCalls: []*funcCall{},
	}
}

func (m *MockEventBus) Handler() go_ddd.HandlerFunc { return nil }

func (m *MockEventBus) DispatchFrom(ctx context.Context, producer go_ddd.EventProducer) error {
	return nil
}

func (m *MockEventBus) Use(middleware go_ddd.MiddlewareFunc) { return }

func (m *MockEventBus) RegisterView(view go_ddd.View) error { return nil }

func (m *MockEventBus) RegisterPolicy(policy go_ddd.Policy) error { return nil }

// ExpectDispatch adds an expectation for a call to Dispatch with specific arguments.
func (m *MockEventBus) ExpectDispatch(ctx context.Context, event go_ddd.Event) {
	m.dispatchCalls = append(m.dispatchCalls, &funcCall{
		aggregateType: event.AggregateType(),
		aggregateID:   event.AggregateID().String(),
		eventType:     event.Type(),
	})
}

// Dispatch simulates the Dispatch function and verifies expectations.
func (m *MockEventBus) Dispatch(ctx context.Context, event go_ddd.Event) error {
	// Check if a matching expectation exists
	res := make([]*funcCall, 0)
	for _, call := range m.dispatchCalls {
		if call.aggregateType == event.AggregateType() &&
			call.aggregateID == event.AggregateID().String() &&
			call.eventType == event.Type() {
			continue // Mock execution, can return a simulated error if needed
		} else {
			res = append(res, call)
		}
	}

	if len(res) < len(m.dispatchCalls) {
		m.dispatchCalls = res
		return nil
	}

	// No matching expectation found
	m.t.Fatalf("Unexpected call to Dispatch: ctx=%v, event=%v", ctx, event)
	return nil // Unreachable, but avoids compilation errors
}

// AssertExpectations verifies that all expected calls to Dispatch were made.
func (m *MockEventBus) AssertExpectations(t *testing.T) {
	if len(m.dispatchCalls) > 0 {
		t.Errorf("Not all expected Dispatch calls were made. %d calls remaining", len(m.dispatchCalls))
		for _, call := range m.dispatchCalls {
			t.Errorf("  - Missing call: aggregateType=%v, eventType=%v", call.aggregateType, call.eventType)
		}
	}
}

type TestPolicy struct {
	subscribedTo []string
}

func (p *TestPolicy) When(event go_ddd.Event) (go_ddd.Command, error) {
	switch event.Type() {
	case "github.com/paulvitic/ddd-go/inMemory.testEventPayload":
		return go_ddd.NewCommand(testCommand{Name: "value"}), nil
	default:
		return nil, errors.New("unknown event type")
	}
}

func (p *TestPolicy) SubscribedTo() []string {
	return p.subscribedTo
}

func translator(from []byte) (go_ddd.Event, error) {
	event, err := go_ddd.EventFromJsonString(string(from))
	if err != nil {
		return nil, err
	}
	return event, nil
}

func TestNewInMemoryMessageConsumer(t *testing.T) {
	queue := make(chan string, 1)
	t.Cleanup(func() {
		close(queue)
	})
	// Test successful creation
	consumer := MessageConsumer(go_ddd.NewMessageConsumer("target", translator), &queue)
	assert.NotNil(t, consumer)

	// Test failure with nil arguments
	assert.Panics(t, func() {
		MessageConsumer(go_ddd.NewMessageConsumer("target", translator), nil)
	})
}

func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := make(chan string, 1)
	t.Cleanup(func() {
		cancel()
		close(queue)
	})

	testEvent := go_ddd.NewEventProducer().
		RegisterEvent("aggType", go_ddd.NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()

	// Create a mock for EventBus with expectations for Dispatch calls
	eventBusMock := NewMockEventBus(t)
	eventBusMock.ExpectDispatch(ctx, testEvent) // Assuming a MockEventBus implementation exists

	consumer := MessageConsumer(go_ddd.NewMessageConsumer("target", translator), &queue)
	consumer.SetEventBus(eventBusMock)

	err := consumer.Start()
	assert.NoError(t, err)

	jsonString, err := testEvent.ToJsonString()
	assert.NoError(t, err)

	// Send a message to the queue
	queue <- jsonString

	// Wait for message processing (adjust timing if needed)
	time.Sleep(300 * time.Millisecond)

	// Verify expected calls to EventBus.Dispatch
	eventBusMock.AssertExpectations(t) // Assuming a MockEventBus assert method
}

func TestStop(t *testing.T) {
	consumer := &messageConsumer{
		running: atomic.Bool{},
	}

	err := consumer.Start()
	assert.NoError(t, err)
	assert.True(t, consumer.running.Load())

	consumer.Stop()
	assert.False(t, consumer.running.Load())
}

// TestInMemoryMessageConsumer is a social test which tests the in-memory message consumer with real components.
func TestInMemoryMessageConsumer(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping test in short mode.")
	//}
	msgQueue := make(chan string, 1)

	wg := sync.WaitGroup{}
	wg.Add(2)
	calls := make([]string, 0)

	middleware := func(next go_ddd.HandlerFunc) go_ddd.HandlerFunc {
		return func(ctx context.Context, msg go_ddd.Payload) (interface{}, error) {
			wg.Done()
			calls = append(calls, "middleware")
			return next(ctx, msg)
		}
	}

	testCommandBus := go_ddd.NewCommandBus()
	handler := func(context.Context, go_ddd.Command) error {
		return nil
	}
	service := go_ddd.NewCommandService(handler, testCommand{})
	err := testCommandBus.RegisterService(service)
	assert.NoError(t, err)

	testEventBus := go_ddd.NewEventBus(testCommandBus)
	testEventBus.Use(middleware)

	err = testEventBus.RegisterPolicy(&TestPolicy{
		[]string{"github.com/paulvitic/ddd-go/inMemory.testEventPayload"},
	})
	assert.NoError(t, err)

	consumer := MessageConsumer(go_ddd.NewMessageConsumer("target", translator), &msgQueue)
	consumer.SetEventBus(testEventBus)
	err = consumer.Start()
	assert.NoError(t, err)

	t.Cleanup(func() {
		close(msgQueue)
	})

	testEvent := go_ddd.NewEventProducer().
		RegisterEvent("aggType", go_ddd.NewID("ID123"), testEventPayload{Name: "value"}).
		GetFirst()

	jsonString, err := testEvent.ToJsonString()
	assert.NoError(t, err)

	go func() { msgQueue <- jsonString }()
	go func() { msgQueue <- jsonString }()

	wg.Wait()
	assert.Equal(t, []string{"middleware", "middleware"}, calls)
}
