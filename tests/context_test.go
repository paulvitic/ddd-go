package ddd_tests

import (
	"testing"

	"github.com/paulvitic/ddd-go"
)

// Tests for Context
func TestNewContext(t *testing.T) {
	context := ddd.NewContext("test-context")

	if context.Name() != "test-context" {
		t.Errorf("Expected context name to be 'test-context', got '%s'", context.Name())
	}

	if context.IsReady() {
		t.Errorf("New context should not be ready")
	}
}

func TestSimpleResourceRegistration(t *testing.T) {

	// Create context and add resources
	context := ddd.NewContext("test").WithResources(
		ddd.NewResource[ddd.Endpoint](TestEndpoint{}),
	)

	if !context.IsReady() {
		t.Errorf("Context should be ready after adding resources")
	}
}
