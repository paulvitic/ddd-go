package ddd

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Container represents the dependency injection container
type Context struct {
	Logger    *Logger `resource:""`
	name      string
	resources map[reflect.Type]map[string]*resource
	mu        sync.RWMutex
	resolving sync.Map
}

// NewContext creates a new Container
func NewContext(name string) *Context {
	context := &Context{
		name:      name,
		resources: make(map[reflect.Type]map[string]*resource),
	}
	context.registerResource(Resource(NewLogger))
	context.AutoWire(context)
	context.Logger.Info("Context %s created", context.name)
	return context
}

func (c *Context) Name() string {
	return c.name
}

func (c *Context) WithResources(resources ...*resource) *Context {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger.Info("Registering resources")
	for _, resource := range resources {
		c.registerResource(resource)
	}
	return c
}

func (c *Context) registerResource(rsc *resource) {
	typ := rsc.Type()

	if _, exists := c.resources[typ]; !exists {
		c.resources[typ] = make(map[string]*resource)
	}

	c.resources[typ][rsc.Name()] = rsc
}

func (c *Context) bindEndpoints(router *mux.Router) error {
	// Create a subrouter for the context
	router = router.PathPrefix("/" + c.Name()).Subrouter()

	// Resolve all resources using the generic method
	resources := ReourcesByType[Endpoint](c)

	// Bind each endpoint
	for name, resource := range resources {
		// contruct the enspoint to get the path and handlers
		if endpoint, err := c.construct(resource); err != nil {
			c.Logger.Error("can not contruct endpoint %s", name)

		} else {
			if endpoint, ok := endpoint.(Endpoint); !ok {
				c.Logger.Error("resource is not of type Endpoint")
			} else {
				path := endpoint.Path()
				handlers := RequestHandlers(endpoint)

				if resource.scope == Request {
					for method, methodName := range handlers {
						// Capture variables for closure
						currentMethod := method
						currentMethodName := methodName

						wrapperHandler := func(w http.ResponseWriter, r *http.Request) {
							key := uuid.New()
							defer c.ClearRequestScoped(resource, key)

							// Resolve endpoint for this specific request
							requestEndpoint, err := c.Resolve(resource.Type(), key)
							if err != nil {
								http.Error(w, fmt.Sprintf("Failed to resolve endpoint: %v", err), http.StatusInternalServerError)
								return
							}
							callHandlerMethod(requestEndpoint, currentMethodName, w, r)
						}

						router.HandleFunc(path, wrapperHandler).Methods(string(currentMethod))
					}
				} else {
					for method, methodName := range handlers {
						// Capture variables for closure
						currentMethod := method
						currentMethodName := methodName

						handler := func(w http.ResponseWriter, r *http.Request) {
							callHandlerMethod(endpoint, currentMethodName, w, r)
						}

						router.HandleFunc(path, handler).Methods(string(currentMethod))
					}
				}
			}
		}
	}

	return nil
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

// Resolve resolves a dependency from the container
func (c *Context) Resolve(typ reflect.Type, options ...any) (any, error) {
	name, key := c.parseResolveOptions(options...)

	if name == "" {
		name = getDefaultName(typ)
	}

	// Check for circular dependencies
	if _, resolving := c.resolving.LoadOrStore(typ, true); resolving {
		return nil, fmt.Errorf("circular dependency detected for type %v", typ)
	}
	defer c.resolving.Delete(typ)

	c.mu.RLock()
	resource, err := c.getResource(typ, name)
	c.mu.RUnlock()

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
	case Request:
		if key == uuid.Nil {
			panic("request scope resource requires a request ID")
		}
		return c.resolveRequest(resource, key)
	default:
		return nil, fmt.Errorf("unknown scope: %v", resource.scope)
	}
}

func (c *Context) parseResolveOptions(options ...any) (string, uuid.UUID) {
	var name string
	var key uuid.UUID

	for _, option := range options {
		switch v := option.(type) {
		case string:
			name = v
		case uuid.UUID:
			key = v
		}
	}

	return name, key
}

func (c *Context) resolveSingleton(info *resource) (any, error) {
	var err error
	info.initOnce.Do(func() {
		var instance any
		instance, err = c.construct(info)
		if err == nil {
			info.instance.Store(instance)
		}
	})

	if err != nil {
		return nil, err
	}

	return info.instance.Load(), nil
}

// Store the instance in the resource's instancePool by a uuid key
// The key represents a request ID, goroutine ID, or other identifier
func (c *Context) resolveRequest(info *resource, key uuid.UUID) (any, error) {

	if instance, ok := info.instancePool.Load(key); ok {
		return instance, nil
	}

	instance, err := c.construct(info)
	if err != nil {
		return nil, err
	}

	info.instancePool.Store(key, instance)
	return instance, nil
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

	if hooks, ok := resource.hooks.(LifecycleHooks[any]); ok {
		if hooks.OnInit != nil {
			if err := hooks.OnInit(instance); err != nil {
				return nil, err
			}
		}
		if hooks.OnStart != nil {
			if err := hooks.OnStart(instance); err != nil {
				return nil, err
			}
		}
	}

	return instance, nil
}

func (c *Context) resolveFactoryParams(factoryType reflect.Type) ([]reflect.Value, error) {
	params := make([]reflect.Value, factoryType.NumIn())
	for i := 0; i < factoryType.NumIn(); i++ {
		paramType := factoryType.In(i)
		param, err := c.Resolve(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d of type %v: %w", i, paramType, err)
		}
		params[i] = reflect.ValueOf(param)
	}
	return params, nil
}

// AutoWire automatically injects dependencies into the fields of the given struct
func (c *Context) AutoWire(target any) error {
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

		dependency, err := c.Resolve(field.Type(), options...)
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
		for _, info := range implementations {
			if hooks, ok := info.hooks.(LifecycleHooks[any]); ok {
				if hooks.OnDestroy != nil {
					instance := info.instance.Load()
					if instance != nil {
						if err := hooks.OnDestroy(instance); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

// ClearRequestScoped clears all request-scoped dependencies
func (c *Context) ClearAllRequestScoped() {
	for _, resources := range c.resources {
		for _, resource := range resources {
			if resource.scope == Request {
				resource.instancePool.Range(func(key, value interface{}) bool {
					c.ClearRequestScoped(resource, key.(uuid.UUID))
					return true
				})
			}
		}
	}
}

// ClearRequestScopedByID clears request-scoped dependencies for a specific ID
func (c *Context) ClearRequestScoped(rsc *resource, id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resource, err := c.getResource(rsc.Type(), rsc.Name())
	if err != nil {
		if instance, exists := resource.instancePool.Load(id); exists {
			// Run destroy hooks if they exist
			if hooks, ok := resource.hooks.(LifecycleHooks[any]); ok {
				if hooks.OnDestroy != nil {
					if err := hooks.OnDestroy(instance); err != nil {
						c.Logger.Error("Error during OnDestroy hook for request-scoped dependency: %v", err)
					}
				}
			}
			// Remove the instance from the pool
			resource.instancePool.Delete(id)
		}
	}
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	runes := []rune(t.Name())
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Generic method to resolve all instances of a specific interface
func ReourcesByType[T any](c *Context) map[string]*resource {
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
	instance, err := c.Resolve(reflect.TypeOf(&t).Elem(), options...)
	if err != nil {
		return t, err
	}
	return instance.(T), nil
}

func AutoWire[T any](c *Context, target *T) error {
	return c.AutoWire(target)
}
