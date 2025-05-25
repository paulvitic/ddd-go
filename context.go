package ddd

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/gorilla/mux"
)

// Container represents the dependency injection container
type Context struct {
	Logger       *Logger `resource:""`
	name         string
	dependencies map[reflect.Type]map[string]*dependencyInfo
	mu           sync.RWMutex
	resolving    sync.Map
}

// dependencyInfo holds information about a registered dependency
type dependencyInfo struct {
	constructor  reflect.Value
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	hooks        any
	instancePool sync.Map
}

// LifecycleHooks defines lifecycle hooks for dependencies
type LifecycleHooks[T any] struct {
	OnInit    func(T) error
	OnStart   func(T) error
	OnDestroy func(T) error
}

// NewContext creates a new Container
func NewContext(name string) *Context {
	context := &Context{
		name:         name,
		dependencies: make(map[reflect.Type]map[string]*dependencyInfo),
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
		if err := c.registerResource(resource); err != nil {
			panic(fmt.Sprintf("failed to register resource %s: %v", resource.name, err))
		}
	}
	return c
}

func (c *Context) registerResource(resource *resource) error {
	typ := resource.Type()

	if _, exists := c.dependencies[typ]; !exists {
		c.dependencies[typ] = make(map[string]*dependencyInfo)
	}

	c.dependencies[typ][resource.Name()] = &dependencyInfo{
		constructor:  reflect.ValueOf(resource.Value()),
		scope:        resource.Scope(),
		hooks:        resource.hooks,
		instancePool: sync.Map{},
	}

	return nil
}

func (c *Context) registerEndpoints(router *mux.Router) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a subrouter for the context
	contextRouter := router.PathPrefix("/" + c.Name()).Subrouter()

	// Iterate through all registered dependencies to find Endpoints
	for typ, implementations := range c.dependencies {
		// Check if the type implements the Endpoint interface
		endpointType := reflect.TypeOf((*Endpoint)(nil)).Elem()
		if typ.Implements(endpointType) || reflect.PtrTo(typ).Implements(endpointType) {
			// Resolve each implementation of this endpoint type
			for name, info := range implementations {
				endpoint, err := c.Resolve(typ, name)
				if err != nil {
					c.Logger.Error("Failed to resolve endpoint %s: %v", name, err)
					continue
				}

				if ep, ok := endpoint.(Endpoint); ok {
					path := ep.Path()
					handlers := RequestHandlers(reflect.TypeOf(endpoint))

					if info.scope == Request {
						// Create wrapper handlers for each discovered method
						for method, methodName := range handlers {
							wrapperHandler := func(w http.ResponseWriter, r *http.Request) {
								// Resolve the endpoint for this specific request (creates new instance)
								requestEndpoint, err := c.Resolve(typ, name)
								if err != nil {
									http.Error(w, fmt.Sprintf("Failed to resolve endpoint: %v", err), http.StatusInternalServerError)
									return
								}

								// Get the handler method and call it
								endpointValue := reflect.ValueOf(requestEndpoint)
								handlerMethod := endpointValue.MethodByName(methodName)

								if !handlerMethod.IsValid() {
									http.Error(w, "Handler method not found", http.StatusInternalServerError)
									return
								}

								// Call the handler method
								handlerMethod.Call([]reflect.Value{
									reflect.ValueOf(w),
									reflect.ValueOf(r),
								})

								// Clean up request-scoped dependencies after the request
								defer c.ClearRequestScoped()
							}
							contextRouter.HandleFunc(path, wrapperHandler).Methods(string(method))
							c.Logger.Info("Registered %s handler (request-scoped) for endpoint %s in context %s", method, path, c.Name())
						}

						// Clean up the temporary instance
						// c.ClearRequestScoped()
					} else {
						endpointValue := reflect.ValueOf(endpoint)

						// Register handlers for each discovered method
						for method, methodName := range handlers {
							handlerMethod := endpointValue.MethodByName(methodName)
							if !handlerMethod.IsValid() {
								continue
							}

							// Create a closure that captures the handler method
							handler := func(w http.ResponseWriter, r *http.Request) {
								// Call the handler method directly
								handlerMethod.Call([]reflect.Value{
									reflect.ValueOf(w),
									reflect.ValueOf(r),
								})
							}
							contextRouter.HandleFunc(path, handler).Methods(string(method))
							c.Logger.Info("Registered %s handler (static) for endpoint %s in context %s", method, path, c.Name())
						}
					}
				} else {
					c.Logger.Warn("Resolved object is not an Endpoint: %s", name)
				}
			}
		}
	}
	return nil
}

// Resolve resolves a dependency from the container
func (c *Context) Resolve(typ reflect.Type, options ...any) (any, error) {
	name := c.getResolveName(options...)

	// Check for circular dependencies
	if _, resolving := c.resolving.LoadOrStore(typ, true); resolving {
		return nil, fmt.Errorf("circular dependency detected for type %v", typ)
	}
	defer c.resolving.Delete(typ)

	c.mu.RLock()
	info, err := c.getDependencyInfo(typ, name)
	c.mu.RUnlock()

	if err != nil {
		return nil, err
	}

	return c.resolveDependency(info)
}

func (c *Context) getResolveName(options ...any) string {
	for _, option := range options {
		if n, ok := option.(string); ok {
			return n
		}
	}
	return ""
}

func (c *Context) getDependencyInfo(typ reflect.Type, name string) (*dependencyInfo, error) {
	implementations, exists := c.dependencies[typ]
	if !exists {
		return nil, fmt.Errorf("no dependency registered for type %v", typ)
	}

	if name == "" {
		name = getDefaultName(typ)
	}

	info, exists := implementations[name]
	if !exists {
		return nil, fmt.Errorf("no dependency named '%s' registered for type %v", name, typ)
	}

	return info, nil
}

func (c *Context) resolveDependency(info *dependencyInfo) (any, error) {
	switch info.scope {
	case Singleton:
		return c.resolveSingleton(info)
	case Prototype:
		return c.construct(info)
	case Request:
		return c.resolveRequest(info)
	default:
		return nil, fmt.Errorf("unknown scope: %v", info.scope)
	}
}

func (c *Context) resolveSingleton(info *dependencyInfo) (any, error) {
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

func (c *Context) resolveRequest(info *dependencyInfo) (any, error) {
	key := getGoroutineID()
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

func (c *Context) construct(info *dependencyInfo) (any, error) {
	params, err := c.resolveConstructorParams(info.constructor.Type())
	if err != nil {
		return nil, err
	}

	results := info.constructor.Call(params)
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	instance := results[0].Interface()

	if hooks, ok := info.hooks.(LifecycleHooks[any]); ok {
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

func (c *Context) resolveConstructorParams(constructorType reflect.Type) ([]reflect.Value, error) {
	params := make([]reflect.Value, constructorType.NumIn())
	for i := 0; i < constructorType.NumIn(); i++ {
		paramType := constructorType.In(i)
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

	for i := 0; i < v.NumField(); i++ {
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

	for _, implementations := range c.dependencies {
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
func (c *Context) ClearRequestScoped() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, implementations := range c.dependencies {
		for _, info := range implementations {
			if info.scope == Request {
				info.instancePool = sync.Map{}
			}
		}
	}
}

// Helper functions
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toCamelCase(t.Name())
}

func getGoroutineID() uint64 {
	return uint64(reflect.ValueOf(make(chan int)).Pointer())
}

func convertToInterfaceFunc(v reflect.Value) func(any) error {
	if v.IsNil() {
		return nil
	}
	return func(i any) error {
		results := v.Call([]reflect.Value{reflect.ValueOf(i)})
		if len(results) == 0 {
			return nil
		}
		err, _ := results[0].Interface().(error)
		return err
	}
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
