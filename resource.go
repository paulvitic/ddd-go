package ddd

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"unicode"
)

type Scope int

const (
	Singleton Scope = iota
	Prototype
	Request
)

// ResourceTag is the struct tag used to specify dependencies
const ResourceTag = "resource"

// Dependency represents a dependency of a resource
type Dependency struct {
	ResourceType string // Name of the field in the struct
	ResourceName string // Custom name for the dependency
}

// Resource represents a resource in the application context
// TODO, right now all resources are Singleton. Could add logic for Prototype and Request scope
type Resource struct {
	name         string
	value        any // the struct
	resourceType string
	constructor  reflect.Value
	dependencies []Dependency
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	hooks        any
	instancePool sync.Map
}

// LifecycleHooks defines lifecycle hooks for resources
type ResourceLifecycleHooks[T any] struct {
	OnInit    func(T) error
	OnStart   func(T) error
	OnDestroy func(T) error
}

func ResourceFromConstructor(constructor any, options ...any) (*Resource, error) {
	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return nil, fmt.Errorf("constructor must be a function")
	}

	if constructorType.NumOut() == 0 || (constructorType.NumOut() == 2 && !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
		return nil, fmt.Errorf("constructor must return (T) or (T, error)")
	}

	typ := constructorType.Out(0)

	name, scope, hooks := processOptions(typ, options...)

	return &Resource{
		name:         name,
		constructor:  reflect.ValueOf(constructor),
		scope:        scope,
		hooks:        hooks,
		instancePool: sync.Map{},
	}, nil
}

// Create a resource from struct to be added to the context to be autowired
// and instatiated as a singelton, prototype or request scoped resource instance
func NewResource[I any](value any, name ...string) *Resource {
	// make sure value is a struct kind type or pointer to struct
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}
	if valueType.Kind() != reflect.Struct {
		panic("value must be a struct type or pointer to struct")
	}

	// name argument is optional. If not provided get the structs type name in camel case format
	var resourceName string
	if len(name) > 0 {
		resourceName = name[0]
	} else {
		// Get the type name and convert to camel case
		typeName := valueType.Name()
		resourceName = toCamelCase(typeName)
	}

	// Get the interface type
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	if interfaceType.Kind() != reflect.Interface {
		panic("type parameter I must be an interface")
	}

	// Verify that the value implements the interface
	if !valueType.Implements(interfaceType) {
		panic("value does not implement the specified interface")
	}

	// Parse dependencies from struct tags
	var dependencies []Dependency
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		fieldName := toCamelCase(field.Name)
		tag := field.Tag.Get(ResourceTag)
		if tag != "" {
			fieldName = tag
		}

		// Use the tag value as the custom resource name
		dependencies = append(dependencies, Dependency{
			ResourceType: fieldName,
			ResourceName: tag,
		})
	}

	return &Resource{
		name:         resourceName,
		resourceType: interfaceType.String(),
		value:        value,
		dependencies: dependencies,
	}
}

func processOptions(typ reflect.Type, options ...any) (string, Scope, any) {
	var name string
	scope := Singleton
	var hooks any

	for _, option := range options {
		switch v := option.(type) {
		case string:
			name = v
		case Scope:
			scope = v
		default:
			if h, ok := isLifecycleHooks(v); ok {
				hooks = h
			}
		}
	}

	if name == "" {
		name = getDefaultName(typ)
	}

	return name, scope, hooks
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toCamelCase(t.Name())
}

func getResourceLifecyleHooks(t reflect.Type) *ResourceLifecycleHooks[any] {
	onInit, exists := t.MethodByName("OnInit")
	onStart, exists := t.MethodByName("OnStart")
	onDestroy, exists := t.MethodByName("OnDestroy")

	return &ResourceLifecycleHooks[any]{
		OnInit:    onInit,
		OnStart:   onStart,
		OnDestroy: onDestroy,
	}

}

func isLifecycleHooks(v any) (ResourceLifecycleHooks[any], bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return ResourceLifecycleHooks[any]{}, false
	}

	rt := rv.Type()
	if rt.NumField() != 3 {
		return ResourceLifecycleHooks[any]{}, false
	}

	onInitField, hasOnInit := rt.FieldByName("OnInit")
	onStartField, hasOnStart := rt.FieldByName("OnStart")
	onDestroyField, hasOnDestroy := rt.FieldByName("OnDestroy")

	if !hasOnInit || !hasOnStart || !hasOnDestroy {
		return ResourceLifecycleHooks[any]{}, false
	}

	isValidHook := func(f reflect.StructField) bool {
		return f.Type.Kind() == reflect.Func &&
			f.Type.NumIn() == 1 &&
			f.Type.NumOut() == 1 &&
			f.Type.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
	}

	if !isValidHook(onInitField) || !isValidHook(onStartField) || !isValidHook(onDestroyField) {
		return ResourceLifecycleHooks[any]{}, false
	}

	return ResourceLifecycleHooks[any]{
		OnInit:    convertToInterfaceFunc(rv.FieldByName("OnInit")),
		OnStart:   convertToInterfaceFunc(rv.FieldByName("OnStart")),
		OnDestroy: convertToInterfaceFunc(rv.FieldByName("OnDestroy")),
	}, true
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

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
