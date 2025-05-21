package ddd_tests

import (
	"fmt"
	"reflect"
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
type SomeStructRepo struct{}

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
}

func TestNewResourceWithCustomOptions(t *testing.T) {
	testCases := []struct {
		name          string
		options       []any
		expectedName  string
		expectedScope ddd.Scope
	}{
		{
			name:          "No options",
			options:       []any{UserRepository{}},
			expectedName:  "userRepository",
			expectedScope: ddd.Singleton,
		},
		{
			name:          "Custom name only",
			options:       []any{"customRepo"},
			expectedName:  "customRepo",
			expectedScope: ddd.Singleton,
		},
		{
			name:          "Custom scope only",
			options:       []any{ddd.Prototype},
			expectedName:  "userRepository",
			expectedScope: ddd.Prototype,
		},
		{
			name:          "Request scope",
			options:       []any{ddd.Request},
			expectedName:  "userRepository",
			expectedScope: ddd.Request,
		},
		{
			name:          "Both custom name and scope",
			options:       []any{"customRepo", ddd.Prototype},
			expectedName:  "customRepo",
			expectedScope: ddd.Prototype,
		},
		{
			name:          "Both custom scope and name (different order)",
			options:       []any{ddd.Request, "customRepo"},
			expectedName:  "customRepo",
			expectedScope: ddd.Request,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resource := ddd.NewResource[Repository](UserRepository{}, tc.options...)
			resource := ddd.NewResource[Repository](tc.options...)

			if resource.Name() != tc.expectedName {
				t.Errorf("Expected name to be '%s', got '%s'", tc.expectedName, resource.Name())
			}

			if resource.Scope() != tc.expectedScope {
				t.Errorf("Expected scope to be %v, got %v", tc.expectedScope, resource.Scope())
			}
		})
	}
}

func TestInterfaceVerification(t *testing.T) {
	// Test successful interface verification
	t.Run("Valid implementation", func(t *testing.T) {
		// This should not panic
		_ = ddd.NewResource[ddd.Logger](&ddd.Logger{})
	})

	t.Run("Value type implementing with pointer receiver", func(t *testing.T) {
		// This should not panic even though SimpleLogger methods have pointer receivers
		_ = ddd.NewResource[ddd.Logger](ddd.Logger{})
	})

	t.Run("Invalid implementation", func(t *testing.T) {
		// Define a recovery function to catch panics
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for interface verification, but it didn't happen")
			}
		}()

		// This should panic because string doesn't implement Logger
		_ = ddd.NewResource[ddd.Logger]("not an implementation")
	})
}

func TestDependencyDetection(t *testing.T) {
	t.Run("Multiple dependencies", func(t *testing.T) {
		serviceResource := ddd.NewResource[Service](UserService{ApiKey: "test123"})
		dependencies := serviceResource.Dependencies()

		if len(dependencies) != 3 {
			t.Errorf("Expected 3 dependencies, got %d", len(dependencies))
		}

		// Check for specific dependencies
		dependencyMap := make(map[string]string)
		for _, dep := range dependencies {
			dependencyMap[dep.ResourceType] = dep.ResourceName
		}

		expectedDeps := map[string]string{
			"logger":   "logger",
			"userRepo": "userRepo",
			"apiKey":   "apiKey",
		}

		for resourceType, resourceName := range expectedDeps {
			if actualName, ok := dependencyMap[resourceType]; !ok || actualName != resourceName {
				t.Errorf("Expected dependency %s with name %s, got name %s", resourceType, resourceName, actualName)
			}
		}
	})

	t.Run("No dependencies", func(t *testing.T) {
		repoResource := ddd.NewResource[Repository](UserRepository{})
		dependencies := repoResource.Dependencies()

		if len(dependencies) != 0 {
			t.Errorf("Expected 0 dependencies, got %d", len(dependencies))
		}
	})

	t.Run("Custom dependency names via tags", func(t *testing.T) {
		// Create a struct with a resource tag
		type StructWithTag struct {
			LogLevel string `resource:"logLevel"`
		}

		// We don't need to implement Logger for this test
		// since we're just checking the dependencies
		loggerResource := ddd.NewResource[any](StructWithTag{})
		dependencies := loggerResource.Dependencies()

		if len(dependencies) != 1 {
			t.Errorf("Expected 1 dependency, got %d", len(dependencies))
		}

		if dependencies[0].ResourceType != "logLevel" || dependencies[0].ResourceName != "logLevel" {
			t.Errorf("Expected dependency type 'logLevel' and name 'logLevel', got type '%s' and name '%s'",
				dependencies[0].ResourceType, dependencies[0].ResourceName)
		}
	})
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

// Test that hooks properly return errors
func TestLifecycleHookErrors(t *testing.T) {
	// Create the resource
	resource := ddd.NewResource[ddd.Logger](ErrorLogger{LogLevel: "DEBUG"})
	hooks := resource.LifecycleHooks()

	// Call the hook and check for error
	err := hooks.OnInit(ErrorLogger{})
	if err == nil || err.Error() != "init error" {
		t.Errorf("Expected OnInit hook to return 'init error', got: %v", err)
	}
}

// Example of how resources would be used
func ExampleNewResource() {
	// Create resources
	loggerResource := ddd.NewResource[ddd.Logger](ddd.Logger{})
	repoResource := ddd.NewResource[Repository](UserRepository{}, "userRepo", ddd.Prototype)
	serviceResource := ddd.NewResource[Service](UserService{ApiKey: "abc123"}, "userService")

	// In a real application, these would be added to a context:
	// appContext := NewContext("app").WithResources(
	//     loggerResource,
	//     repoResource,
	//     serviceResource,
	// )

	// For the example, just print the resource types
	fmt.Printf("Created resources: %s, %s, %s\n",
		reflect.TypeOf(loggerResource).String(),
		reflect.TypeOf(repoResource).String(),
		reflect.TypeOf(serviceResource).String())
	// Output: Created resources: *ddd.Resource, *ddd.Resource, *ddd.Resource
}
