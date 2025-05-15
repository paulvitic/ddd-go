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

// LifecycleHooks defines lifecycle hooks for resources
type LifecycleHooks[T any] struct {
	OnInit    func(T) error
	OnStart   func(T) error
	OnDestroy func(T) error
}

// Resource represents a resource in the application context
type Resource struct {
	name         string
	value        any // the struct
	resourceType string
	dependencies []Dependency
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	hooks        *LifecycleHooks[any]
	instancePool sync.Map
}

// Create a resource from struct to be added to the context to be autowired
// and instatiated as a singelton, prototype or request scoped resource instance
func NewResource[I any](value any, options ...any) *Resource {
	// make sure value is a struct kind type or pointer to struct
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}
	if valueType.Kind() != reflect.Struct {
		panic("value must be a struct type or pointer to struct")
	}

	// Get the interface type
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	if interfaceType.Kind() != reflect.Interface {
		panic("type parameter I must be an interface")
	}

	// Verify that the value implements the interface
	// FIXED: Check both pointer and value types
	implementsInterface := valueType.Implements(interfaceType) || reflect.PtrTo(valueType).Implements(interfaceType)
	if !implementsInterface {
		panic("value does not implement the specified interface")
	}

	resourceName, scope := processOptions(valueType, options)

	// Parse dependencies from struct tags
	dependencies := parseDependencies(valueType)

	lifecycleHooks := getLifecycleHooks(value, valueType)

	return &Resource{
		name:         resourceName,
		resourceType: interfaceType.String(),
		value:        value,
		scope:        scope,
		dependencies: dependencies,
		hooks:        lifecycleHooks,
	}
}

func (r *Resource) Name() string {
	return r.name
}

func (r *Resource) Type() string {
	return r.resourceType
}

func (r *Resource) Scope() Scope {
	return r.scope
}

func (r *Resource) Dependencies() []Dependency {
	return r.dependencies
}

func (r *Resource) LifecycleHooks() *LifecycleHooks[any] {
	return r.hooks
}

func parseDependencies(valueType reflect.Type) []Dependency {
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
	return dependencies
}

func processOptions(typ reflect.Type, options ...any) (string, Scope) {
	var name string
	scope := Singleton

	for _, option := range options {
		switch v := option.(type) {
		case string:
			name = v
		case Scope:
			scope = v
		default:
			panic(fmt.Sprintf("unexpected option type: %T", option))
		}
	}

	if name == "" {
		name = getDefaultName(typ)
	}

	return name, scope
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toCamelCase(t.Name())
}

func getLifecycleHooks(value any, valueType reflect.Type) *LifecycleHooks[any] {
	// We need the value as a pointer for method calls
	valueValue := reflect.ValueOf(value)

	// If it's not a pointer, create a pointer to it
	if valueValue.Kind() != reflect.Ptr {
		newValue := reflect.New(valueType)
		newValue.Elem().Set(valueValue)
		valueValue = newValue
	}

	// The type we look for methods on must be the pointer type
	methodType := valueValue.Type()

	// Initialize with nil functions
	hooks := &LifecycleHooks[any]{
		OnInit:    nil,
		OnStart:   nil,
		OnDestroy: nil,
	}

	// Check for OnInit method
	if onInit, exists := methodType.MethodByName("OnInit"); exists {
		hooks.OnInit = func(a any) error {
			// We need to use the actual instance, not valueValue
			instanceValue := reflect.ValueOf(a)

			// If it's not a pointer but the method has a pointer receiver, wrap it
			if instanceValue.Kind() != reflect.Ptr && onInit.Type.In(0).Kind() == reflect.Ptr {
				newInstance := reflect.New(instanceValue.Type())
				newInstance.Elem().Set(instanceValue)
				instanceValue = newInstance
			}

			results := onInit.Func.Call([]reflect.Value{instanceValue})
			if len(results) > 0 && !results[0].IsNil() {
				return results[0].Interface().(error)
			}
			return nil
		}
	}

	// Check for OnStart method
	if onStart, exists := methodType.MethodByName("OnStart"); exists {
		hooks.OnStart = func(a any) error {
			// We need to use the actual instance, not valueValue
			instanceValue := reflect.ValueOf(a)

			// If it's not a pointer but the method has a pointer receiver, wrap it
			if instanceValue.Kind() != reflect.Ptr && onStart.Type.In(0).Kind() == reflect.Ptr {
				newInstance := reflect.New(instanceValue.Type())
				newInstance.Elem().Set(instanceValue)
				instanceValue = newInstance
			}

			results := onStart.Func.Call([]reflect.Value{instanceValue})
			if len(results) > 0 && !results[0].IsNil() {
				return results[0].Interface().(error)
			}
			return nil
		}
	}

	// Check for OnDestroy method
	if onDestroy, exists := methodType.MethodByName("OnDestroy"); exists {
		hooks.OnDestroy = func(a any) error {
			// We need to use the actual instance, not valueValue
			instanceValue := reflect.ValueOf(a)

			// If it's not a pointer but the method has a pointer receiver, wrap it
			if instanceValue.Kind() != reflect.Ptr && onDestroy.Type.In(0).Kind() == reflect.Ptr {
				newInstance := reflect.New(instanceValue.Type())
				newInstance.Elem().Set(instanceValue)
				instanceValue = newInstance
			}

			results := onDestroy.Func.Call([]reflect.Value{instanceValue})
			if len(results) > 0 && !results[0].IsNil() {
				return results[0].Interface().(error)
			}
			return nil
		}
	}

	return hooks
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// func ResourceFromConstructor(constructor any, options ...any) (*Resource, error) {
// 	constructorType := reflect.TypeOf(constructor)
// 	if constructorType.Kind() != reflect.Func {
// 		return nil, fmt.Errorf("constructor must be a function")
// 	}

// 	if constructorType.NumOut() == 0 || (constructorType.NumOut() == 2 && !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
// 		return nil, fmt.Errorf("constructor must return (T) or (T, error)")
// 	}

// 	typ := constructorType.Out(0)

// 	name, scope := processOptions(typ, options...)

// 	return &Resource{
// 		name:         name,
// 		constructor:  reflect.ValueOf(constructor),
// 		scope:        scope,
// 		hooks:        hooks,
// 		instancePool: sync.Map{},
// 	}, nil
// }

// func isLifecycleHooks(v any) (ResourceLifecycleHooks[any], bool) {
// 	rv := reflect.ValueOf(v)
// 	if rv.Kind() != reflect.Struct {
// 		return ResourceLifecycleHooks[any]{}, false
// 	}

// 	rt := rv.Type()
// 	if rt.NumField() != 3 {
// 		return ResourceLifecycleHooks[any]{}, false
// 	}

// 	onInitField, hasOnInit := rt.FieldByName("OnInit")
// 	onStartField, hasOnStart := rt.FieldByName("OnStart")
// 	onDestroyField, hasOnDestroy := rt.FieldByName("OnDestroy")

// 	if !hasOnInit || !hasOnStart || !hasOnDestroy {
// 		return ResourceLifecycleHooks[any]{}, false
// 	}

// 	isValidHook := func(f reflect.StructField) bool {
// 		return f.Type.Kind() == reflect.Func &&
// 			f.Type.NumIn() == 1 &&
// 			f.Type.NumOut() == 1 &&
// 			f.Type.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
// 	}

// 	if !isValidHook(onInitField) || !isValidHook(onStartField) || !isValidHook(onDestroyField) {
// 		return ResourceLifecycleHooks[any]{}, false
// 	}

// 	return ResourceLifecycleHooks[any]{
// 		OnInit:    convertToInterfaceFunc(rv.FieldByName("OnInit")),
// 		OnStart:   convertToInterfaceFunc(rv.FieldByName("OnStart")),
// 		OnDestroy: convertToInterfaceFunc(rv.FieldByName("OnDestroy")),
// 	}, true
// }

// func convertToInterfaceFunc(v reflect.Value) func(any) error {
// 	if v.IsNil() {
// 		return nil
// 	}
// 	return func(i any) error {
// 		results := v.Call([]reflect.Value{reflect.ValueOf(i)})
// 		if len(results) == 0 {
// 			return nil
// 		}
// 		err, _ := results[0].Interface().(error)
// 		return err
// 	}
// }
