package ddd

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// Define test interfaces
type Logger interface {
	Log(message string)
}

type Service interface {
	DoSomething() error
}

type Repository interface {
	FindById(id int) string
}

// Define implementations
type SimpleLogger struct {
	LogLevel string `resource:"logLevel"`
}

func (l *SimpleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", l.LogLevel, message)
}

func (l *SimpleLogger) OnInit() error {
	return nil
}

func (l *SimpleLogger) OnDestroy() error {
	return nil
}

// FullLifecycleLogger implements all lifecycle hooks
type FullLifecycleLogger struct {
	SimpleLogger
}

func (l *FullLifecycleLogger) OnInit() error    { return nil }
func (l *FullLifecycleLogger) OnStart() error   { return nil }
func (l *FullLifecycleLogger) OnDestroy() error { return nil }

// TestLogger has flags to check if hooks were called
type TestLogger struct {
	LogLevel      string `resource:"logLevel"`
	InitCalled    bool
	StartCalled   bool
	DestroyCalled bool
}

func (l *TestLogger) Log(message string) {
	// No-op
}

func (l *TestLogger) OnInit() error {
	l.InitCalled = true
	return nil
}

func (l *TestLogger) OnStart() error {
	l.StartCalled = true
	return nil
}

func (l *TestLogger) OnDestroy() error {
	l.DestroyCalled = true
	return nil
}

// ErrorLogger returns errors from lifecycle hooks
type ErrorLogger struct {
	LogLevel string `resource:"logLevel"`
}

func (l *ErrorLogger) Log(message string) {
	// No-op
}

func (l *ErrorLogger) OnInit() error {
	return errors.New("init error")
}

// NoHooksStruct doesn't implement any lifecycle hooks
type NoHooksStruct struct{}

func (s *NoHooksStruct) DoSomething() error { return nil }

type UserService struct {
	Logger     Logger     `resource:"logger"`
	Repository Repository `resource:"userRepo"`
	ApiKey     string     `resource:"apiKey"`
}

func (s *UserService) DoSomething() error {
	s.Logger.Log("Doing something with repository")
	s.Repository.FindById(1)
	return nil
}

func (s *UserService) OnInit() error {
	if s.ApiKey == "" {
		return errors.New("ApiKey cannot be empty")
	}
	return nil
}

func (s *UserService) OnStart() error {
	return nil
}

type UserRepository struct{}

func (r *UserRepository) FindById(id int) string {
	return fmt.Sprintf("User with ID %d", id)
}

func (r *UserRepository) OnInit() error {
	return nil
}

// Test cases
func TestNewResourceBasicProperties(t *testing.T) {
	// Test with default name and scope
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})

	if loggerResource.Name() != "simpleLogger" {
		t.Errorf("Expected name to be 'simpleLogger', got '%s'", loggerResource.Name())
	}

	if loggerResource.Scope() != Singleton {
		t.Errorf("Expected default scope to be Singleton, got %v", loggerResource.Scope())
	}
}

func TestNewResourceWithCustomOptions(t *testing.T) {
	testCases := []struct {
		name          string
		options       []any
		expectedName  string
		expectedScope Scope
	}{
		{
			name:          "No options",
			options:       []any{},
			expectedName:  "userRepository",
			expectedScope: Singleton,
		},
		{
			name:          "Custom name only",
			options:       []any{"customRepo"},
			expectedName:  "customRepo",
			expectedScope: Singleton,
		},
		{
			name:          "Custom scope only",
			options:       []any{Prototype},
			expectedName:  "userRepository",
			expectedScope: Prototype,
		},
		{
			name:          "Request scope",
			options:       []any{Request},
			expectedName:  "userRepository",
			expectedScope: Request,
		},
		{
			name:          "Both custom name and scope",
			options:       []any{"customRepo", Prototype},
			expectedName:  "customRepo",
			expectedScope: Prototype,
		},
		{
			name:          "Both custom scope and name (different order)",
			options:       []any{Request, "customRepo"},
			expectedName:  "customRepo",
			expectedScope: Request,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource := NewResource[Repository](UserRepository{}, tc.options...)

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
		_ = NewResource[Logger](&SimpleLogger{LogLevel: "DEBUG"})
	})

	t.Run("Value type implementing with pointer receiver", func(t *testing.T) {
		// This should not panic even though SimpleLogger methods have pointer receivers
		_ = NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})
	})

	t.Run("Invalid implementation", func(t *testing.T) {
		// Define a recovery function to catch panics
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for interface verification, but it didn't happen")
			}
		}()

		// This should panic because string doesn't implement Logger
		_ = NewResource[Logger]("not an implementation")
	})
}

func TestDependencyDetection(t *testing.T) {
	t.Run("Multiple dependencies", func(t *testing.T) {
		serviceResource := NewResource[Service](UserService{ApiKey: "test123"})
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
		repoResource := NewResource[Repository](UserRepository{})
		dependencies := repoResource.Dependencies()

		if len(dependencies) != 0 {
			t.Errorf("Expected 0 dependencies, got %d", len(dependencies))
		}
	})

	t.Run("Custom dependency names via tags", func(t *testing.T) {
		loggerResource := NewResource[Logger](SimpleLogger{})
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
		resource := NewResource[Logger](FullLifecycleLogger{})
		hooks := resource.LifecycleHooks()

		if hooks.OnInit == nil || hooks.OnStart == nil || hooks.OnDestroy == nil {
			t.Errorf("Expected all hooks to be detected, got OnInit=%v, OnStart=%v, OnDestroy=%v",
				hooks.OnInit != nil, hooks.OnStart != nil, hooks.OnDestroy != nil)
		}
	})

	t.Run("Some lifecycle hooks", func(t *testing.T) {
		// UserService has OnInit and OnStart but no OnDestroy
		resource := NewResource[Service](UserService{})
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
		resource := NewResource[Service](NoHooksStruct{})
		hooks := resource.LifecycleHooks()

		if hooks.OnInit != nil || hooks.OnStart != nil || hooks.OnDestroy != nil {
			t.Errorf("Expected no hooks to be detected, got OnInit=%v, OnStart=%v, OnDestroy=%v",
				hooks.OnInit != nil, hooks.OnStart != nil, hooks.OnDestroy != nil)
		}
	})
}

func TestValuePointerReceiver(t *testing.T) {
	t.Run("Value type with pointer receiver methods", func(t *testing.T) {
		valueLogger := SimpleLogger{LogLevel: "DEBUG"}
		resource := NewResource[Logger](valueLogger)
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
		pointerLogger := &SimpleLogger{LogLevel: "DEBUG"}
		resource := NewResource[Logger](pointerLogger)
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
		scope         Scope
		expectedScope Scope
	}{
		{"Default scope", Scope(-1), Singleton}, // Using -1 to represent no specified scope
		{"Singleton scope", Singleton, Singleton},
		{"Prototype scope", Prototype, Prototype},
		{"Request scope", Request, Request},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var resource *Resource
			if tc.scope == Scope(-1) {
				// No scope specified, use default
				resource = NewResource[Logger](SimpleLogger{})
			} else {
				resource = NewResource[Logger](SimpleLogger{}, tc.scope)
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
		_ = NewResource[Logger](SimpleLogger{}, 123)
	})
}

// Test that lifecycle hooks can actually be called
func TestLifecycleHookExecution(t *testing.T) {
	// Create the resource
	logger := &TestLogger{LogLevel: "DEBUG"}
	resource := NewResource[Logger](logger)
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
	resource := NewResource[Logger](ErrorLogger{LogLevel: "DEBUG"})
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
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})
	repoResource := NewResource[Repository](UserRepository{}, "userRepo", Prototype)
	serviceResource := NewResource[Service](UserService{ApiKey: "abc123"}, "userService")

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
