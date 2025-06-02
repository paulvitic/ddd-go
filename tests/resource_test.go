package ddd_tests

import (
	"reflect"
	"testing"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/repository"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/infrastructure/adapter/file"
)

// Test cases
func TestStructReturn(t *testing.T) {
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

func TestRepositoryStereotypeReturn(t *testing.T) {
	// Test with default name and scope
	repoResource := ddd.Resource(file.NewUserRepository)

	if repoResource.Name() != "userRepository" {
		t.Errorf("Expected name to be 'userRepository', got '%s'", repoResource.Name())
	}

	if repoResource.Types()[0] != reflect.TypeOf((*repository.UserRepository)(nil)).Elem() {
		t.Errorf(
			"Expected resource type to be '%s', got '%s'",
			reflect.TypeOf((*repository.UserRepository)(nil)).Elem(), repoResource.Types()[0])
	}

	if repoResource.Scope() != ddd.Singleton {
		t.Errorf("Expected default scope to be Singleton, got %v", repoResource.Scope())
	}
}
