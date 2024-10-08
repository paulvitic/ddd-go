package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockCommandTranslator(r *http.Request) (ddd.Command, error) {
	// Create a valid command here
	cmd := ddd.NewCommand(struct{ Name string }{Name: "test"})
	return cmd, nil
}

type mockCommandBus struct {
	dispatchCalled bool
	dispatchErr    error
}

func (m *mockCommandBus) Use(middleware ddd.MiddlewareFunc) {
	return
}

func (m *mockCommandBus) Dispatch(ctx context.Context, cmd ddd.Command) error {
	m.dispatchCalled = true
	return m.dispatchErr
}

func (m *mockCommandBus) RegisterService(service ddd.CommandService) error {
	return nil
}

func TestCommandEndpoint(t *testing.T) {
	e := NewCommandEndpoint("/test", []string{http.MethodPost, http.MethodPut, http.MethodDelete}, mockCommandTranslator)
	assert.Equal(t, "/test", e.Path())

	mockBus := &mockCommandBus{}
	e.RegisterCommandBus(mockBus)

	// Create a payload
	payload := map[string]string{
		"key": "value",
	}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	// Create a new request with the payload
	req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(payloadBytes))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := e.Handler()
	handler(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
}
