package ddd

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// Context represents the application context which manages resources
type Context struct {
	name            string
	resources       []*Resource
	resourcesByType map[string][]*Resource // map of resource type to resources
	resourcesByName map[string]*Resource   // map of resource name to resource
	mutex           sync.RWMutex
	ready           bool
}

// NewContext creates a new application context with the given name
func NewContext(name string) *Context {
	return &Context{
		name:            name,
		resources:       make([]*Resource, 0),
		resourcesByType: make(map[string][]*Resource),
		resourcesByName: make(map[string]*Resource),
	}
}

// WithResources adds resources to the context and returns the context
func (c *Context) WithResources(resources ...*Resource) *Context {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Add resources to the context
	for _, resource := range resources {
		c.resources = append(c.resources, resource)

		// Add to type map
		c.resourcesByType[resource.Type()] = append(c.resourcesByType[resource.Type()], resource)

		// Add to name map
		c.resourcesByName[resource.Name()] = resource
	}

	// Sort resources by dependency count
	c.sortResourcesByDependencyCount()

	// Initialize resources
	c.initializeResources()

	c.ready = true

	return c
}

// ResourcesByType returns all resources of the given type
func (c *Context) ResourcesByType(resourceType string) ([]any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	resources, exists := c.resourcesByType[resourceType]
	if !exists || len(resources) == 0 {
		return nil, false
	}

	result := make([]any, 0, len(resources))
	for _, resource := range resources {
		instance, ok := c.getInstance(resource)
		if ok {
			result = append(result, instance)
		}
	}

	if len(result) == 0 {
		return nil, false
	}

	return result, true
}

// ResourceByName returns the resource with the given name
func (c *Context) ResourceByName(name string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	resource, exists := c.resourcesByName[name]
	if !exists {
		return nil, false
	}

	return c.getInstance(resource)
}

// ResourceByTypeAndName returns the resource with the given type and name
func (c *Context) ResourceByTypeAndName(resourceType, name string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	resources, exists := c.resourcesByType[resourceType]
	if !exists {
		return nil, false
	}

	for _, resource := range resources {
		if resource.Name() == name {
			return c.getInstance(resource)
		}
	}

	return nil, false
}

// Endpoints returns all resources that implement the Endpoint interface
func (c *Context) Endpoints() []Endpoint {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Look for resources with Endpoint type
	var endpoints []Endpoint

	// Check all resources
	for _, resource := range c.resources {
		// Check if the resource type matches Endpoint interface
		if resource.Type() == "ddd.Endpoint" {
			instance, ok := c.getInstance(resource)
			if !ok || instance == nil {
				continue
			}

			// Type assertion with nil check
			endpoint, ok := instance.(Endpoint)
			if ok {
				endpoints = append(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

// sortResourcesByDependencyCount sorts the resources by dependency count in ascending order
func (c *Context) sortResourcesByDependencyCount() {
	sort.Slice(c.resources, func(i, j int) bool {
		return len(c.resources[i].Dependencies()) < len(c.resources[j].Dependencies())
	})
}

// getInstance returns the instance of the resource based on its scope
func (c *Context) getInstance(resource *Resource) (any, bool) {
	switch resource.Scope() {
	case Singleton:
		// For singletons, we should have already created an instance
		instance := resource.instance.Load()
		if instance != nil {
			return instance, true
		}
		return nil, false

	case Prototype:
		// For prototypes, create a new instance each time
		instance, err := c.createInstance(resource)
		if err != nil {
			return nil, false
		}
		return instance, true

	case Request:
		// For request scope, we would typically use a request ID
		// but since we don't have one here, we'll just create a new instance
		instance, err := c.createInstance(resource)
		if err != nil {
			return nil, false
		}
		return instance, true

	default:
		return nil, false
	}
}

// createInstance creates a new instance of the resource
func (c *Context) createInstance(resource *Resource) (any, error) {
	// Clone the original value
	value := reflect.ValueOf(resource.value)
	if value.Kind() == reflect.Ptr {
		// If it's a pointer, create a new instance of the pointed type
		if value.IsNil() {
			return nil, fmt.Errorf("cannot create instance from nil pointer")
		}
		newInstance := reflect.New(value.Elem().Type())
		newInstance.Elem().Set(value.Elem())
		value = newInstance
	} else {
		// If it's a value, create a new instance of the same type
		newInstance := reflect.New(value.Type()).Elem()
		newInstance.Set(value)
		value = newInstance
	}

	// Resolve dependencies
	if err := c.resolveDependencies(value); err != nil {
		return nil, err
	}

	// Convert to interface
	instance := value.Interface()

	// Call OnInit hook if available
	if resource.hooks != nil && resource.hooks.OnInit != nil {
		if err := resource.hooks.OnInit(instance); err != nil {
			return nil, err
		}
	}

	return instance, nil
}

// resolveDependencies resolves the dependencies of the resource
func (c *Context) resolveDependencies(value reflect.Value) error {
	// If it's a pointer, we want to work with the pointed value
	valueElem := value
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return fmt.Errorf("cannot resolve dependencies on nil pointer")
		}
		valueElem = value.Elem()
	}

	// Get the type of the struct
	valueType := valueElem.Type()

	// Iterate through each field in the struct
	for i := 0; i < valueElem.NumField(); i++ {
		field := valueElem.Field(i)
		fieldType := valueType.Field(i)

		// Check if the field has a resource tag
		tag := fieldType.Tag.Get(ResourceTag)
		if tag != "" {
			// This field is a dependency, resolve it

			// Make sure the field is settable (i.e., not unexported)
			if !field.CanSet() {
				return fmt.Errorf("field %s with resource tag is not settable (probably unexported)", fieldType.Name)
			}

			// Check if there's a resource with the same type and name
			depType := field.Type().String()
			var depInstance any
			var found bool

			// First try by type and name
			depInstance, found = c.ResourceByTypeAndName(depType, tag)

			// If not found, try by type
			if !found {
				instances, found := c.ResourcesByType(depType)
				if found && len(instances) > 0 {
					depInstance = instances[0] // Use the first instance
					found = true
				}
			}

			if !found {
				return fmt.Errorf("dependency not found: %s (%s)", tag, depType)
			}

			// Null check
			if depInstance == nil {
				return fmt.Errorf("dependency instance is nil: %s (%s)", tag, depType)
			}

			// Set the field value
			depValue := reflect.ValueOf(depInstance)

			// Make sure the field is settable
			if field.CanSet() {
				field.Set(depValue)
			} else {
				return fmt.Errorf("cannot set field %s", fieldType.Name)
			}
		}
	}

	return nil
}

// initializeResources initializes all resources in dependency order
func (c *Context) initializeResources() error {
	for _, resource := range c.resources {
		// Skip if it's not a singleton
		if resource.Scope() != Singleton {
			continue
		}

		// Create an instance of the resource
		instance, err := c.createInstance(resource)
		if err != nil {
			return err
		}

		// Store the instance
		resource.instance.Store(instance)

		// Call OnStart hook if available
		if resource.hooks != nil && resource.hooks.OnStart != nil {
			if err := resource.hooks.OnStart(instance); err != nil {
				return err
			}
		}
	}

	return nil
}

// Start calls the OnStart method for all resources that haven't been started yet
func (c *Context) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, resource := range c.resources {
		// Skip if it's not a singleton or if it doesn't have an OnStart hook
		if resource.Scope() != Singleton || resource.hooks == nil || resource.hooks.OnStart == nil {
			continue
		}

		// Get the instance
		instance := resource.instance.Load()
		if instance == nil {
			continue
		}

		// Call OnStart hook
		if err := resource.hooks.OnStart(instance); err != nil {
			return err
		}
	}

	return nil
}

// Destroy calls the OnDestroy method for all resources
func (c *Context) Destroy() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Iterate in reverse order to destroy in reverse dependency order
	for i := len(c.resources) - 1; i >= 0; i-- {
		resource := c.resources[i]

		// Skip if it's not a singleton or if it doesn't have an OnDestroy hook
		if resource.Scope() != Singleton || resource.hooks == nil || resource.hooks.OnDestroy == nil {
			continue
		}

		// Get the instance
		instance := resource.instance.Load()
		if instance == nil {
			continue
		}

		// Call OnDestroy hook
		if err := resource.hooks.OnDestroy(instance); err != nil {
			return err
		}
	}

	return nil
}

// Name returns the name of the context
func (c *Context) Name() string {
	return c.name
}

// IsReady returns true if the context is ready
func (c *Context) IsReady() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ready
}
