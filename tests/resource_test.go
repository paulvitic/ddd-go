package ddd_tests

import (
	"testing"

	"github.com/paulvitic/ddd-go"
)

// Test cases
func TestNewResourceBasicProperties(t *testing.T) {
	// Test with default name and scope
	loggerResource := ddd.NewResource[ddd.Logger]()

	if loggerResource.Name() != "logger" {
		t.Errorf("Expected name to be 'logger', got '%s'", loggerResource.Name())
	}

	if loggerResource.Type() != "ddd.Logger" {
		t.Errorf("Expected resource type to be 'ddd.logger', got '%s'", loggerResource.Type())
	}

	if loggerResource.Scope() != ddd.Singleton {
		t.Errorf("Expected default scope to be Singleton, got %v", loggerResource.Scope())
	}
}

type SomeStruct struct{}
type SomeDependencyInterface interface{}
type SomeDependencyStruct struct{}
type SomeStructRepo struct {
	logger                  ddd.Logger              `resource:""`
	someDependencyInterface SomeDependencyInterface `resource:""`
	someDependencyStruct    SomeDependencyStruct    `resource:"customDepenedencyName"`
	someDependencyPointer   *SomeDependencyStruct   `resource:""`
	nonResourceDependency   string
}

func (s SomeStructRepo) Save(aggregate *SomeStruct) error    { return nil }
func (s SomeStructRepo) Load(id ddd.ID) (*SomeStruct, error) { return nil, nil }
func (s SomeStructRepo) Delete(id ddd.ID) error              { return nil }
func (s SomeStructRepo) Update(aggregate *SomeStruct) error  { return nil }

func TestNewResourceInterfaceDeclaration(t *testing.T) {

	repoResource := ddd.NewResource[ddd.Repository[SomeStruct]](SomeStructRepo{}, "customRepoName")

	if repoResource.Name() != "customRepoName" {
		t.Errorf("Expected name to be 'logger', got '%s'", repoResource.Name())
	}

	if repoResource.Type() != "ddd.Repository[SomeStruct]" {
		t.Errorf("Expected resource type to be 'ddd.Repository[SomeStruct]', got '%s'", repoResource.Type())
	}

	if repoResource.Scope() != ddd.Singleton {
		t.Errorf("Expected default scope to be Singleton, got %v", repoResource.Scope())
	}

	if len(repoResource.Dependencies()) != 4 {
		t.Errorf("Expected 4 resolved resource dependencies, got %v", len(repoResource.Dependencies()))
	}

	if repoResource.Dependencies()[2].ResourceName != "customDepenedencyName" {
		t.Errorf("Expected dependency name 'customDepenedencyName', got %s", repoResource.Dependencies()[1].ResourceName)
	}

	if !repoResource.Dependencies()[3].IsPointer {
		t.Errorf("Expected dependency to be a pointer")
	}
}

func TestLifecycleHooks(t *testing.T) {
	t.Run("All lifecycle hooks", func(t *testing.T) {
		resource := ddd.NewResource[ddd.Logger](ddd.Logger{})
		hooks := resource.LifecycleHooks()

		if hooks.OnInit == nil || hooks.OnStart == nil || hooks.OnDestroy == nil {
			t.Errorf("Expected all hooks to be detected, got OnInit=%v, OnStart=%v, OnDestroy=%v",
				hooks.OnInit != nil, hooks.OnStart != nil, hooks.OnDestroy != nil)
		}
	})

	t.Run("Some lifecycle hooks", func(t *testing.T) {
		// UserService has OnInit and OnStart but no OnDestroy
		resource := ddd.NewResource[Service](UserService{})
		hooks := resource.LifecycleHooks()

		if hooks.OnInit == nil {
			t.Errorf("Expected OnInit hook to be detected")
		}

		if hooks.OnStart == nil {
			t.Errorf("Expected OnStart hook to be detected")
		}

		if hooks.OnDestroy != nil {
			t.Errorf("Expected no OnDestroy hook to be detected")
		}
	})

	t.Run("No lifecycle hooks", func(t *testing.T) {
		resource := ddd.NewResource[Service](NoHooksStruct{})
		hooks := resource.LifecycleHooks()

		if hooks.OnInit != nil || hooks.OnStart != nil || hooks.OnDestroy != nil {
			t.Errorf("Expected no hooks to be detected, got OnInit=%v, OnStart=%v, OnDestroy=%v",
				hooks.OnInit != nil, hooks.OnStart != nil, hooks.OnDestroy != nil)
		}
	})
}

func TestValuePointerReceiver(t *testing.T) {
	t.Run("Value type with pointer receiver methods", func(t *testing.T) {
		valueLogger := ddd.Logger{}
		resource := ddd.NewResource[ddd.Logger](valueLogger)
		hooks := resource.LifecycleHooks()

		// These should work even though SimpleLogger has pointer receiver methods
		if hooks.OnInit == nil {
			t.Errorf("Expected OnInit hook to be detected for value type")
		}

		if hooks.OnDestroy == nil {
			t.Errorf("Expected OnDestroy hook to be detected for value type")
		}
	})

	t.Run("Pointer type with pointer receiver methods", func(t *testing.T) {
		pointerLogger := &ddd.Logger{}
		resource := ddd.NewResource[ddd.Logger](pointerLogger)
		hooks := resource.LifecycleHooks()

		if hooks.OnInit == nil {
			t.Errorf("Expected OnInit hook to be detected for pointer type")
		}

		if hooks.OnDestroy == nil {
			t.Errorf("Expected OnDestroy hook to be detected for pointer type")
		}
	})
}

func TestResourceScopes(t *testing.T) {
	testCases := []struct {
		name          string
		scope         ddd.Scope
		expectedScope ddd.Scope
	}{
		{"Default scope", ddd.Scope(-1), ddd.Singleton}, // Using -1 to represent no specified scope
		{"Singleton scope", ddd.Singleton, ddd.Singleton},
		{"Prototype scope", ddd.Prototype, ddd.Prototype},
		{"Request scope", ddd.Request, ddd.Request},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var resource *ddd.Resource
			if tc.scope == ddd.Scope(-1) {
				// No scope specified, use default
				resource = ddd.NewResource[ddd.Logger](ddd.Logger{})
			} else {
				resource = ddd.NewResource[ddd.Logger](ddd.Logger{}, tc.scope)
			}

			if resource.Scope() != tc.expectedScope {
				t.Errorf("Expected scope %v, got %v", tc.expectedScope, resource.Scope())
			}
		})
	}
}

func TestInvalidOptions(t *testing.T) {
	t.Run("Invalid option type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for invalid option type, but it didn't happen")
			}
		}()

		// This should panic because 123 is not a valid option type
		_ = ddd.NewResource[ddd.Logger](ddd.Logger{}, 123)
	})
}

// Test that lifecycle hooks can actually be called
func TestLifecycleHookExecution(t *testing.T) {
	// Create the resource
	logger := &TestLogger{LogLevel: "DEBUG"}
	resource := ddd.NewResource[ddd.Logger](logger)
	hooks := resource.LifecycleHooks()

	// Call the hooks
	if err := hooks.OnInit(logger); err != nil {
		t.Errorf("OnInit hook returned error: %v", err)
	}

	if err := hooks.OnStart(logger); err != nil {
		t.Errorf("OnStart hook returned error: %v", err)
	}

	if err := hooks.OnDestroy(logger); err != nil {
		t.Errorf("OnDestroy hook returned error: %v", err)
	}

	// Verify flags were set
	if !logger.InitCalled {
		t.Errorf("OnInit hook was not called")
	}

	if !logger.StartCalled {
		t.Errorf("OnStart hook was not called")
	}

	if !logger.DestroyCalled {
		t.Errorf("OnDestroy hook was not called")
	}
}
