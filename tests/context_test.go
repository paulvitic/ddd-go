package ddd_tests

import (
	"testing"

	"github.com/paulvitic/ddd-go"
)

// Tests for Context
func TestNewContext(t *testing.T) {
	context := ddd.NewContext("test-context").
		WithResources(
			ddd.Resource(NewTestEndpoint, ddd.Request),
		)

	if context.Name() != "test-context" {
		t.Errorf("Expected context name to be 'test-context', got '%s'", context.Name())
	}
}
