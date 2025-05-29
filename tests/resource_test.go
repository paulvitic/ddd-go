package ddd_tests

import (
	"reflect"
	"testing"

	"github.com/paulvitic/ddd-go"
)

// Test cases
func TestNewResourceBasicProperties(t *testing.T) {
	// Test with default name and scope
	loggerResource := ddd.Resource(ddd.NewLogger)

	if loggerResource.Name() != "logger" {
		t.Errorf("Expected name to be 'logger', got '%s'", loggerResource.Name())
	}

	if loggerResource.Types()[0] != reflect.TypeOf((*ddd.Logger)(nil)) {
		t.Errorf(
			"Expected resource type to be '%s', got '%s'",
			reflect.TypeOf((*ddd.Logger)(nil)).Elem(), loggerResource.Types()[0])
	}

	if loggerResource.Scope() != ddd.Singleton {
		t.Errorf("Expected default scope to be Singleton, got %v", loggerResource.Scope())
	}
}
