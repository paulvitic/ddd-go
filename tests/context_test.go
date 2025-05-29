package ddd_tests

import (
	"testing"

	"github.com/paulvitic/ddd-go"
)

// Tests for Context
func TestNewContext(t *testing.T) {
	context := ddd.NewContext("test").
		WithResources(
			ddd.Resource(NewTestEndpoint),
		)

	if context.Name() != "test" {
		t.Errorf("Expected context name to be 'test', got '%s'", context.Name())
	}
}
