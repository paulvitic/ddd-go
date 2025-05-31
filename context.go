package ddd

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/gorilla/mux"
)

type contextKey string

const AppContextKey contextKey = "appContext"

type ContextFactory = func(parent context.Context) *Context

// Container represents the dependency injection container
type Context struct {
	context.Context
	logger    *Logger
	name      string
	resources map[reflect.Type]map[string]*resource
	resolving sync.Map
	mu        sync.RWMutex
}

// NewContext creates a new Container
func NewContext(parentCtx context.Context, name string) *Context {
	loggerResource := Resource(NewLogger)
	logger, _ := loggerResource.Create(nil)

	context := &Context{
		Context:   parentCtx,
		logger:    logger.(*Logger),
		name:      name,
		resources: make(map[reflect.Type]map[string]*resource),
	}
	context.logger.Info("%s context created", context.name)

	context.WithResources(
		loggerResource,
		Resource(NewEventBus),
	)

	return context
}

func (c *Context) Name() string {
	return c.name
}

func (c *Context) WithResources(resources ...*resource) *Context {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, resource := range resources {
		c.registerResource(resource)
	}

	return c
}

func (c *Context) Logger() *Logger {
	return c.logger
}

func (c *Context) registerResource(rsc *resource) {
	types := rsc.Types()
	registeredTypes := make([]string, 0)

	for _, typ := range types {
		if _, exists := c.resources[typ]; !exists {
			c.resources[typ] = make(map[string]*resource)
		}

		c.resources[typ][rsc.Name()] = rsc
		registeredTypes = append(registeredTypes, ResourceTypeName(typ))
	}
	c.logger.Info("resource registered for type(s) %s", strings.Join(registeredTypes, ", "))
}

func (c *Context) bindEndpoints(router *mux.Router) error {
	// Resolve all resources using the generic method
	resources := ResourcesByType[Endpoint](c)

	// Bind each endpoint
	for name, resource := range resources {
		// contruct the endpoint to get the path and handlers
		if endpoint, err := c.construct(resource); err != nil {
			c.logger.Error("can not contruct endpoint %s", name)
		} else {
			if endpoint, ok := endpoint.(Endpoint); !ok {
				c.logger.Error("resource is not of type Endpoint")
			} else {
				BindEndpoint(endpoint, router)
			}
		}
	}

	return nil
}

// resolve resolves a dependency from the container
func (c *Context) resolve(typ reflect.Type, options ...any) (any, error) {
	name := c.parseResolveOptions(options...)

	if name == "" {
		name = getDefaultName(typ)
	}

	// Check for circular dependencies
	if _, resolving := c.resolving.LoadOrStore(typ, true); resolving {
		return nil, fmt.Errorf("circular dependency detected for type %v", typ)
	}
	defer c.resolving.Delete(typ)

	resource, err := c.getResource(typ, name)

	if err != nil {
		return nil, err
	}

	switch resource.scope {
	case Singleton:
		// For singletons, we should have already created an instance
		return c.resolveSingleton(resource)
	case Prototype:
		// For prototypes, create a new instance each time
		return c.construct(resource)
	default:
		return nil, fmt.Errorf("unknown scope: %v", resource.scope)
	}
}

// autoWire automatically injects dependencies into the fields of a given struct
// if the field is tagged with 'resource' and has public accessibility
func (c *Context) autoWire(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		tag, exists := t.Field(i).Tag.Lookup("resource")

		if !exists {
			continue
		}

		var options []any
		if tag != "" {
			options = append(options, tag)
		}

		dependency, err := c.resolve(field.Type(), options...)
		if err != nil {
			return fmt.Errorf("failed to autowire field %s: %w", t.Field(i).Name, err)
		}

		field.Set(reflect.ValueOf(dependency))
	}

	return nil
}

func (c *Context) Destroy() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, implementations := range c.resources {
		for _, resource := range implementations {
			instance := resource.instance.Load()
			if instance != nil {
				if err := resource.ExecuteLifecycleHook(instance, "OnDestroy"); err != nil {
					return fmt.Errorf("OnDestroy hook failed for %s: %w", ResourceTypeName(resource.Types()[0]), err)
				}
			}
		}
	}
	return nil
}

func (c *Context) parseResolveOptions(options ...any) string {
	var name string

	for _, option := range options {
		switch v := option.(type) {
		case string:
			name = v
		default:
			name = ""
		}
	}

	return name
}

// gets resource by type and name
func (c *Context) getResource(typ reflect.Type, name string) (*resource, error) {
	resources, exists := c.resources[typ]
	if !exists {
		return nil, fmt.Errorf("no dependency registered for type %v", typ)
	}

	if name == "" {
		name = getDefaultName(typ)
	}

	resource, exists := resources[name]
	if !exists {
		return nil, fmt.Errorf("no dependency named '%s' registered for type %v", name, typ)
	}

	return resource, nil
}

func (c *Context) resolveSingleton(resource *resource) (any, error) {
	var err error
	resource.initOnce.Do(func() {
		var instance any
		instance, err = c.construct(resource)
		if err == nil {
			resource.instance.Store(instance)
		}
	})

	if err != nil {
		return nil, err
	}

	return resource.instance.Load(), nil
}

func (c *Context) construct(resource *resource) (any, error) {
	params, err := c.resolveFactoryParams(resource.factory.Type())
	if err != nil {
		return nil, err
	}

	results := resource.factory.Call(params)
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	instance := results[0].Interface()

	// AutoWire dependencies after construction
	if err := c.autoWire(instance); err != nil {
		return nil, fmt.Errorf("failed to autowire dependencies: %w", err)
	}

	// Execute lifecycle hooks in order: OnInit -> OnStart
	if err := resource.ExecuteLifecycleHook(instance, "OnInit"); err != nil {
		return nil, fmt.Errorf("OnInit hook failed: %w", err)
	}

	if err := resource.ExecuteLifecycleHook(instance, "OnStart"); err != nil {
		return nil, fmt.Errorf("OnStart hook failed: %w", err)
	}

	return instance, nil
}

func (c *Context) resolveFactoryParams(factoryType reflect.Type) ([]reflect.Value, error) {
	params := make([]reflect.Value, factoryType.NumIn())
	for i := range factoryType.NumIn() {
		paramType := factoryType.In(i)
		param, err := c.resolve(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d of type %v: %w", i, paramType, err)
		}
		params[i] = reflect.ValueOf(param)
	}
	return params, nil
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	runes := []rune(t.Name())
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ==========================================================
// Generic functions
// ==========================================================
// to resolve all instances of a specific interface
func ResourcesByType[T any](c *Context) map[string]*resource {
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	results := make(map[string]*resource)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for typ, resources := range c.resources {
		// Check if the registered type implements or matches the target type
		if typ.AssignableTo(targetType) || typ == targetType {
			for name, resource := range resources {
				results[name] = resource
			}
		}
	}

	return results
}

func Resolve[T any](c *Context, options ...any) (T, error) {
	var t T
	instance, err := c.resolve(reflect.TypeOf(&t).Elem(), options...)
	if err != nil {
		return t, err
	}
	return instance.(T), nil
}

func AutoWire[T any](c *Context, target *T) error {
	return c.autoWire(target)
}
