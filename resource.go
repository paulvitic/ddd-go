package ddd

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

type Scope int

const (
	Singleton Scope = iota
	Prototype
)

var stereotypes = []reflect.Type{
	reflect.TypeOf((*Endpoint)(nil)).Elem(),
	reflect.TypeOf((*CommandHandler)(nil)).Elem(),
	reflect.TypeOf((*EventHandler)(nil)).Elem(),
}

type resource struct {
	factory      reflect.Value
	types        []reflect.Type
	name         string
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	instancePool sync.Map
}

func Resource(factory any, options ...any) *resource {
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		panic("factory must be a function")
	}

	if factoryType.NumOut() == 0 || (factoryType.NumOut() == 2 && !factoryType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
		panic("factory must return (T) or (T, error)")
	}

	returnType := factoryType.Out(0)
	name, scope := processOptions(returnType, options...)

	// Build the types slice: concrete type first, then stereotype interfaces
	types := []reflect.Type{returnType}

	for _, stereotype := range stereotypes {
		if returnType.Implements(stereotype) {
			types = append(types, stereotype)
		}
	}

	return &resource{
		factory:      reflect.ValueOf(factory),
		types:        types,
		name:         name,
		scope:        scope,
		instancePool: sync.Map{},
	}
}

func (r *resource) Factory() any {
	return r.factory
}

func (r *resource) Name() string {
	return r.name
}

// Types returns all types (concrete + stereotype interfaces)
func (r *resource) Types() []reflect.Type {
	return r.types
}

// HasType checks if the resource implements a specific type
func (r *resource) HasType(t reflect.Type) bool {
	for _, resourceType := range r.types {
		if resourceType == t {
			return true
		}
	}
	return false
}

func (r *resource) Scope() Scope {
	return r.scope
}

func (r *resource) Create(params []reflect.Value) (any, error) {
	switch r.scope {
	case Singleton:
		var err error
		r.initOnce.Do(func() {
			var instance any
			instance, err = r.construct(params)
			if err == nil {
				r.instance.Store(instance)
			}
		})

		if err != nil {
			return nil, err
		}

		return r.instance.Load(), nil
	case Prototype:
		// For prototypes, create a new instance each time
		return r.construct(params)
	default:
		return nil, errors.New(fmt.Sprintf("Scope %w not supported", r.scope))
	}
}

func (r *resource) construct(params []reflect.Value) (any, error) {
	results := r.factory.Call(params)
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	instance := results[0].Interface()
	return instance, nil
}

// ExecuteLifecycleHook discovers and executes a specific lifecycle hook on an instance
func (r *resource) ExecuteLifecycleHook(instance any, methodName string) error {
	instanceValue := reflect.ValueOf(instance)

	method := instanceValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil // Hook doesn't exist, that's fine
	}

	// Verify the method signature
	if !isLifecycleHook(method.Type()) {
		return nil // Invalid signature, skip
	}

	// Call the method
	results := method.Call([]reflect.Value{})

	// Handle return value
	if len(results) == 0 {
		return nil // void method
	}

	if len(results) == 1 {
		if results[0].IsNil() {
			return nil
		}
		return results[0].Interface().(error)
	}

	return nil
}

// isLifecycleHook checks if a method has the correct lifecycle hook signature
func isLifecycleHook(methodType reflect.Type) bool {
	// Should have no parameters (just the receiver)
	if methodType.NumIn() != 0 {
		return false
	}

	// Should return only error or nothing
	if methodType.NumOut() == 0 {
		return true // void methods are acceptable
	}

	if methodType.NumOut() == 1 {
		// Check if it returns error
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		return methodType.Out(0).Implements(errorType)
	}

	return false
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
		}
	}

	if name == "" {
		name = getDefaultName(typ)
	}

	return name, scope
}

func ResourceTypeName(typ reflect.Type) string {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// If it has a name, use it
	if name := typ.Name(); name != "" {
		return name
	}

	// Fallback to string representation
	return typ.String()
}
