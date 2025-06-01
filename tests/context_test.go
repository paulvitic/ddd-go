package ddd_tests

import (
	"context"
	"testing"

	"github.com/gorilla/mux"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context"
)

// Tests for Context
func TestNewContext(t *testing.T) {
	ctx := context.Background()
	router := mux.NewRouter()
	context := test_context.TestContext(ctx, router)
	if context.Name() != "test" {
		t.Errorf("Expected context name to be 'test', got '%s'", context.Name())
	}
}
