package ddd

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type contextKey string

const AppContextKey contextKey = "appContext"

type ContextFactory = func(parent context.Context, router *mux.Router) *Context

// Container represents the dependency injection container
type Context struct {
	context.Context
	name      string
	logger    *Logger
	router    *mux.Router
	eventBus  *EventBus
	resources map[reflect.Type]map[string]*resource
	resolving sync.Map
	mu        sync.RWMutex
}

// NewContext creates a new Container
func NewContext(parentCtx context.Context, router *mux.Router, name string) *Context {
	newCtx := &Context{
		Context:   parentCtx,
		name:      name,
		logger:    NewLogger(),
		resources: make(map[reflect.Type]map[string]*resource),
	}

	newCtx.logger.Info("%s context created", newCtx.name)

	ctxRouter := router.PathPrefix("/" + newCtx.name).Subrouter()
	// Apply middleware to inject context into ALL routes
	ctxRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := context.WithValue(r.Context(), AppContextKey, newCtx)
			r = r.WithContext(reqCtx)
			// Call next handler
			next.ServeHTTP(w, r)
		})
	})
	newCtx.router = ctxRouter

	newCtx.eventBus = NewEventBus(newCtx)

	return newCtx
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

	c.init()
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

		typeName := typ.Name()
		if typeName == "" {
			typeName = typ.Elem().Name()
		}
		registeredTypes = append(registeredTypes, typeName)
	}
	c.logger.Info("registered %s for type(s) %s", rsc.Name(), strings.Join(registeredTypes, ", "))
}

func (c *Context) init() error {

	c.eventBus.Init()

	for _, resourceTypes := range c.resources {
		for _, resource := range resourceTypes {
			switch resource.scope {
			case Singleton:
				// For singletons, we should have already created an instance
				if _, err := c.resolveSingleton(resource); err != nil {
					return err
				}
			case Prototype:
				// For prototypes, create a new instance each time
				if _, err := c.construct(resource); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown scope: %v", resource.scope)
			}
		}
	}
	return nil
}

// resolve resolves a dependency from the container
func (c *Context) resolve(typ reflect.Type, options ...any) (any, error) {

	if typ == reflect.TypeOf(c) {
		return c, nil
	}

	if typ == reflect.TypeOf(c.router) {
		return c.router, nil
	}

	if typ == reflect.TypeOf(c.eventBus) {
		return c.eventBus, nil
	}

	if typ == reflect.TypeOf(c.logger) {
		return c.logger, nil
	}

	name := c.parseResolveOptions(options...)

	if name == "" {
		name = ResourceName(typ)
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
				if err := ExecuteLifecycleHook(instance, "OnDestroy"); err != nil {
					// return fmt.Errorf("OnDestroy hook failed for %s: %w", ResourceTypeName(resource.Types()[0]), err)
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
		// name = ResourceTypeName(typ)
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
	if err := ExecuteLifecycleHook(instance, "OnInit"); err != nil {
		return nil, fmt.Errorf("OnInit hook failed: %w", err)
	}

	// TODO: Not here but with Context start
	// if err := ExecuteLifecycleHook(instance, "OnStart"); err != nil {
	// 	return nil, fmt.Errorf("OnStart hook failed: %w", err)
	// }

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

// ==========================================================
// Generic functions
// ==========================================================
// to resolve all instances of a specific interface
func ResourcesByType[T any](c *Context) map[string]*resource {
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	results := make(map[string]*resource)

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

func ResolveAll[T any](c *Context) ([]T, error) {
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	results := make([]T, 0)
	for typ, resources := range c.resources {
		// Check if the registered type implements or matches the target type
		if typ.AssignableTo(targetType) || typ == targetType {
			for _, resource := range resources {
				switch resource.scope {
				case Singleton:
					// For singletons, we should have already created an instance
					instance, err := c.resolveSingleton(resource)
					if err == nil {
						results = append(results, instance.(T))
					}
				case Prototype:
					// For prototypes, create a new instance each time
					instance, err := c.construct(resource)
					if err == nil {
						results = append(results, instance.(T))
					}
				default:
					return nil, fmt.Errorf("unknown scope: %v", resource.scope)
				}
			}
		}
	}

	return results, nil
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
