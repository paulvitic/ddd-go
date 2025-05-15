package ddd

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type Context struct {
	name      string
	resources []*Resource
	instances map[string]map[string]any
	// from auto_wired
	dependencies map[reflect.Type]map[string]*Resource
	mu           sync.RWMutex
	resolving    sync.Map
}

func NewContext(name string) *Context {
	return &Context{
		name:      name,
		resources: make([]*Resource, 0),
		instances: make(map[string]map[string]any),
		// from auto_wired
		dependencies: make(map[reflect.Type]map[string]*Resource),
	}
}

func (c *Context) Name() string {
	return c.name
}

func (c *Context) Endpoints() []Endpoint {
	// FIXME return inititilizes Endpoint type
	return make([]Endpoint, 0)
}

func (c *Context) WithResources(resources []*Resource) *Context {
	// Add resources to resources slice
	c.resources = append(c.resources, resources...)

	// Initialize resources in dependency order
	c.initializeResources()

	return c
}

func (c *Context) Start() {
	// FIXME start all resources that have a Start() method
}

func (c *Context) Stop() {
	// FIXME stop all resources that have a Stop() method
}

func (c *Context) initializeResources() {
	// Sort resources by dependency count (least dependencies first)
	sort.Slice(c.resources, func(i, j int) bool {
		return len(c.resources[i].Dependencies) < len(c.resources[j].Dependencies)
	})

	// Initialize resources in multiple passes
	for len(c.resources) > 0 {
		initializedInThisPass := false
		var remainingResources []*Resource

		// Try to initialize each remaining resource
		for _, resource := range c.resources {
			// Try to create the instance - this will panic if dependencies aren't available
			func() {
				defer func() {
					if r := recover(); r != nil {
						// If creation failed, add to remaining resources
						remainingResources = append(remainingResources, resource)
						return
					}
					// If we get here, creation succeeded
					initializedInThisPass = true
				}()

				// Create and initialize instance
				instance := c.createInstance(resource)
				if init, ok := instance.(ResourceType); ok {
					// FIXME if resource has init methos call it.
					init.OnInit()
				}

				// Store the instance
				if _, exists := c.instances[resource.ResourceType]; !exists {
					c.instances[resource.ResourceType] = make(map[string]any)
				}
				c.instances[resource.ResourceType][resource.ResourceName] = instance
			}()
		}

		// If we couldn't initialize any resources in this pass, we have a problem
		if !initializedInThisPass {
			// Report which resources couldn't be initialized
			var remainingNames []string
			for _, resource := range remainingResources {
				remainingNames = append(remainingNames, fmt.Sprintf("%s:%s", resource.ResourceType, resource.ResourceName))
			}
			panic(fmt.Sprintf("Cannot initialize remaining resources: %v", remainingNames))
		}

		// Update resources for next pass
		c.resources = remainingResources
	}
}

func (c *Context) createInstance(resource *Resource) any {
	// Create a new instance of the resource
	instance := reflect.New(reflect.TypeOf(resource.Value)).Interface()

	// Autowire dependencies
	value := reflect.ValueOf(instance).Elem()
	for _, dep := range resource.Dependencies {
		field := value.FieldByName(dep.FieldName)
		if !field.IsValid() {
			panic(fmt.Sprintf("Field %s not found in resource %s", dep.FieldName, resource.ResourceName))
		}

		// First try to find an instance with matching interface and resource name
		var depInstance any
		for _, instances := range c.instances {
			if instance, exists := instances[dep.ResourceName]; exists {
				// Check if the instance implements the required interface
				if reflect.TypeOf(instance).Implements(field.Type()) {
					depInstance = instance
					break
				}
			}
		}

		// If not found, try to find any instance implementing the interface
		if depInstance == nil {
			for _, instances := range c.instances {
				for _, instance := range instances {
					if reflect.TypeOf(instance).Implements(field.Type()) {
						depInstance = instance
						break
					}
				}
				if depInstance != nil {
					break
				}
			}
		}

		if depInstance == nil {
			panic(fmt.Sprintf("No suitable dependency found for field %s in resource %s", dep.FieldName, resource.ResourceName))
		}

		// Set the dependency
		field.Set(reflect.ValueOf(depInstance))
	}

	return instance
}

func (c *Context) Register(constructor any, options ...any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function")
	}

	if constructorType.NumOut() == 0 || (constructorType.NumOut() == 2 && !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
		return fmt.Errorf("constructor must return (T) or (T, error)")
	}

	typ := constructorType.Out(0)
	name, scope, hooks := c.processOptions(typ, options...)

	if _, exists := c.dependencies[typ]; !exists {
		c.dependencies[typ] = make(map[string]*Resource)
	}

	c.dependencies[typ][name] = &Resource{
		constructor:  reflect.ValueOf(constructor),
		scope:        scope,
		hooks:        hooks,
		instancePool: sync.Map{},
	}

	return nil
}

// from auto_wired
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

func (c *Context) getDependencyInfo(typ reflect.Type, name string) (*Resource, error) {
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

func (c *Context) resolveDependency(info *Resource) (any, error) {
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

func (c *Context) resolveSingleton(info *Resource) (any, error) {
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

func (c *Context) resolveRequest(info *Resource) (any, error) {
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

func (c *Context) construct(info *Resource) (any, error) {
	params, err := c.resolveConstructorParams(info.constructor.Type())
	if err != nil {
		return nil, err
	}

	results := info.constructor.Call(params)
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	instance := results[0].Interface()

	if hooks, ok := info.hooks.(ResourceLifecycleHooks[any]); ok {
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

		tag := t.Field(i).Tag.Get("autowire")

		if tag == "-" {
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
			if hooks, ok := info.hooks.(ResourceLifecycleHooks[any]); ok {
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

func getGoroutineID() uint64 {
	return uint64(reflect.ValueOf(make(chan int)).Pointer())
}

// Type-safe wrappers

func Register[T any](c *Context, constructor any, options ...any) error {
	return c.Register(constructor, options...)
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
