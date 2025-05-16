package ddd

import (
	"fmt"
	"net/http"
	"testing"
)

// Tests for Context
func TestNewContext(t *testing.T) {
	context := NewContext("test-context")

	if context.Name() != "test-context" {
		t.Errorf("Expected context name to be 'test-context', got '%s'", context.Name())
	}

	if context.IsReady() {
		t.Errorf("New context should not be ready")
	}
}

func TestSimpleResourceRegistration(t *testing.T) {
	// Create resources
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})

	// Create context and add resources
	context := NewContext("test-context").WithResources(loggerResource)

	if !context.IsReady() {
		t.Errorf("Context should be ready after adding resources")
	}

	// Get the resource by type
	instances, found := context.ResourcesByType("ddd.Logger")

	if !found {
		t.Errorf("Resource not found by type")
	}

	if len(instances) != 1 {
		t.Errorf("Expected to find 1 instance, got %d", len(instances))
	}

	logger, ok := instances[0].(Logger)
	if !ok {
		t.Errorf("Expected instance to be a Logger")
	}

	if logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}
}

func TestDependencyResolution(t *testing.T) {
	// Create resources with dependencies
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})
	repoResource := NewResource[Repository](UserRepository{})
	dbServiceResource := NewResource[DatabaseService](&SimpleDatabaseService{ConnectionString: "test-connection"})

	// Create the controller that depends on the above resources
	controllerResource := NewResource[UserController](&SimpleUserController{})

	// Create context and add resources - add them in non-dependency order to test sorting
	context := NewContext("test-context").WithResources(
		controllerResource,
		loggerResource,
		repoResource,
		dbServiceResource,
	)

	// Get the controller
	instances, found := context.ResourcesByType("ddd.UserController")

	if !found {
		t.Errorf("Controller not found by type")
	}

	if len(instances) != 1 {
		t.Errorf("Expected to find 1 instance, got %d", len(instances))
	}

	controller, ok := instances[0].(*SimpleUserController)
	if !ok {
		t.Errorf("Expected instance to be a *SimpleUserController")
	}

	if controller == nil {
		t.Errorf("Expected controller to be non-nil")
	}

	// Check that dependencies are properly resolved
	if controller.Logger == nil {
		t.Errorf("Expected Logger dependency to be resolved")
	}

	if controller.Repository == nil {
		t.Errorf("Expected Repository dependency to be resolved")
	}

	if controller.DbService == nil {
		t.Errorf("Expected DatabaseService dependency to be resolved")
	}

	// Check for proper type and functionality
	if _, ok := controller.Logger.(*SimpleLogger); !ok {
		t.Errorf("Expected Logger to be *SimpleLogger")
	}

	if _, ok := controller.Repository.(*UserRepository); !ok {
		t.Errorf("Expected Repository to be *UserRepository")
	}

	if _, ok := controller.DbService.(*SimpleDatabaseService); !ok {
		t.Errorf("Expected DatabaseService to be *SimpleDatabaseService")
	}
}

func TestContextLifecycleHooks(t *testing.T) {
	// Create resources with hooks
	dbServiceResource := NewResource[DatabaseService](&SimpleDatabaseService{ConnectionString: "test-connection"})
	controllerResource := NewResource[UserController](&SimpleUserController{})
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"})
	repoResource := NewResource[Repository](UserRepository{})

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		controllerResource,
		loggerResource,
		repoResource,
		dbServiceResource,
	)

	// Get the controller
	instances, found := context.ResourcesByType("ddd.UserController")

	if !found {
		t.Errorf("Controller not found by type")
	}

	if len(instances) != 1 {
		t.Errorf("Expected to find 1 instance, got %d", len(instances))
	}

	controller, ok := instances[0].(*SimpleUserController)
	if !ok {
		t.Errorf("Expected instance to be a *SimpleUserController")
	}

	// Check that OnInit was called
	if !controller.InitCalled {
		t.Errorf("Expected OnInit to be called")
	}

	// Call Start and check that OnStart was called
	err := context.Start()
	if err != nil {
		t.Errorf("Expected no error on Start, got %v", err)
	}

	if !controller.StartCalled {
		t.Errorf("Expected OnStart to be called")
	}

	// Call Destroy and check that OnDestroy was called
	err = context.Destroy()
	if err != nil {
		t.Errorf("Expected no error on Destroy, got %v", err)
	}

	if !controller.DestroyCalled {
		t.Errorf("Expected OnDestroy to be called")
	}
}

func TestContextResourceScopes(t *testing.T) {
	// Create resources with different scopes
	singletonLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "SINGLETON"})
	prototypeLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "PROTOTYPE"}, "prototypeLogger", Prototype)
	requestLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "REQUEST"}, "requestLogger", Request)

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		singletonLoggerResource,
		prototypeLoggerResource,
		requestLoggerResource,
	)

	// Get singleton instance twice and compare
	instance1, _ := context.ResourceByName("simpleLogger")
	instance2, _ := context.ResourceByName("simpleLogger")

	// They should be the same instance
	if instance1 != instance2 {
		t.Errorf("Expected singleton instances to be the same")
	}

	// Get prototype instance twice and compare
	instance1, _ = context.ResourceByName("prototypeLogger")
	instance2, _ = context.ResourceByName("prototypeLogger")

	// They should be different instances
	if instance1 == instance2 {
		t.Errorf("Expected prototype instances to be different")
	}

	// Get request instance twice and compare
	instance1, _ = context.ResourceByName("requestLogger")
	instance2, _ = context.ResourceByName("requestLogger")

	// They should be different instances
	if instance1 == instance2 {
		t.Errorf("Expected request instances to be different")
	}
}

func TestGetByName(t *testing.T) {
	// Create resources with custom names
	loggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"}, "customLogger")
	repoResource := NewResource[Repository](UserRepository{}, "customRepo")

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		loggerResource,
		repoResource,
	)

	// Get by name
	instance, found := context.ResourceByName("customLogger")

	if !found {
		t.Errorf("Resource not found by name")
	}

	logger, ok := instance.(Logger)
	if !ok {
		t.Errorf("Expected instance to be a Logger")
	}

	if logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}

	// Get by name that doesn't exist
	_, found = context.ResourceByName("nonExistentResource")

	if found {
		t.Errorf("Should not find non-existent resource")
	}
}

func TestGetByTypeAndName(t *testing.T) {
	// Create multiple resources of the same type with different names
	debugLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"}, "debugLogger")
	infoLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "INFO"}, "infoLogger")

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		debugLoggerResource,
		infoLoggerResource,
	)

	// Get by type and name
	instance, found := context.ResourceByTypeAndName("ddd.Logger", "infoLogger")

	if !found {
		t.Errorf("Resource not found by type and name")
	}

	logger, ok := instance.(Logger)
	if !ok {
		t.Errorf("Expected instance to be a Logger")
	}

	if logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}

	// Get by type and name that doesn't exist
	_, found = context.ResourceByTypeAndName("ddd.Logger", "nonExistentLogger")

	if found {
		t.Errorf("Should not find non-existent resource")
	}
}

func TestMultipleInstancesOfSameType(t *testing.T) {
	// Create multiple resources of the same type with different names
	debugLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"}, "debugLogger")
	infoLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "INFO"}, "infoLogger")

	// Create a service that depends on a logger
	type LoggingService struct {
		Logger Logger `resource:"debugLogger"` // Specifically request the debug logger
	}

	serviceResource := NewResource[any](&LoggingService{})

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		debugLoggerResource,
		infoLoggerResource,
		serviceResource,
	)

	// Get the service
	instances, found := context.ResourcesByType("any")

	if !found {
		t.Errorf("Service not found by type")
	}

	if len(instances) != 1 {
		t.Errorf("Expected to find 1 instance, got %d", len(instances))
	}

	service, ok := instances[0].(*LoggingService)
	if !ok {
		t.Errorf("Expected instance to be a *LoggingService")
	}

	// Check that the dependency is resolved to the correct logger
	if service.Logger == nil {
		t.Errorf("Expected Logger dependency to be resolved")
	}

	// Verify it's the debug logger by checking its log level
	simpleLogger, ok := service.Logger.(*SimpleLogger)
	if !ok {
		t.Errorf("Expected Logger to be *SimpleLogger")
	}

	if simpleLogger.LogLevel != "DEBUG" {
		t.Errorf("Expected LogLevel to be DEBUG, got %s", simpleLogger.LogLevel)
	}
}

func TestGetMultipleResourcesByType(t *testing.T) {
	// Create multiple resources of the same type with different names
	debugLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "DEBUG"}, "debugLogger")
	infoLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "INFO"}, "infoLogger")
	warnLoggerResource := NewResource[Logger](SimpleLogger{LogLevel: "WARN"}, "warnLogger")

	// Create context and add resources
	context := NewContext("test-context").WithResources(
		debugLoggerResource,
		infoLoggerResource,
		warnLoggerResource,
	)

	// Get all loggers by type
	instances, found := context.ResourcesByType("ddd.Logger")

	if !found {
		t.Errorf("Loggers not found by type")
	}

	if len(instances) != 3 {
		t.Errorf("Expected 3 logger instances, got %d", len(instances))
	}

	// Verify we can get each logger and they have the expected log levels
	logLevels := make(map[string]bool)
	for _, instance := range instances {
		logger, ok := instance.(*SimpleLogger)
		if !ok {
			t.Errorf("Expected instance to be a *SimpleLogger")
			continue
		}

		logLevels[logger.LogLevel] = true
	}

	// Check that we have all the expected log levels
	for _, level := range []string{"DEBUG", "INFO", "WARN"} {
		if !logLevels[level] {
			t.Errorf("Expected to find logger with level %s", level)
		}
	}
}

func TestCircularDependencies(t *testing.T) {
	// This test is to demonstrate what happens with circular dependencies
	// In a real implementation, we'd want to detect and handle them gracefully

	// Define two types that depend on each other
	type ServiceA struct {
		B any `resource:"serviceB"`
	}

	type ServiceB struct {
		A any `resource:"serviceA"`
	}

	serviceAResource := NewResource[any](&ServiceA{}, "serviceA")
	serviceBResource := NewResource[any](&ServiceB{}, "serviceB")

	// Create context and add resources - this would ideally detect the circular dependency
	context := NewContext("test-context").WithResources(
		serviceAResource,
		serviceBResource,
	)

	// Try to get service A - this might cause a stack overflow or other issues
	// depending on how circular dependencies are handled
	_, found := context.ResourceByName("serviceA")

	// We're not making any assertions here since the behavior depends on the implementation
	// This is more of a demonstration test
	if !found {
		t.Logf("ServiceA not found, possibly due to circular dependency")
	}
}

// Test endpoints
func TestEndpoints(t *testing.T) {
	// Create a test endpoint
	testEndpoint := &TestEndpoint{
		PathValue: "/test",
		HandlersMap: map[HttpMethod]http.HandlerFunc{
			GET: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "GET test")
			},
			POST: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "POST test")
			},
		},
	}

	// Create a resource for the endpoint
	endpointResource := NewResource[Endpoint](testEndpoint)

	// Create context and add resources
	context := NewContext("test-context").WithResources(endpointResource)

	// Get all endpoints
	endpoints := context.Endpoints()

	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}

	if len(endpoints) > 0 {
		if endpoints[0].Path() != "/test" {
			t.Errorf("Expected endpoint path to be '/test', got '%s'", endpoints[0].Path())
		}

		handlers := endpoints[0].Handlers()
		if _, ok := handlers[GET]; !ok {
			t.Errorf("Expected GET handler to be registered")
		}
		if _, ok := handlers[POST]; !ok {
			t.Errorf("Expected POST handler to be registered")
		}
	}
}
