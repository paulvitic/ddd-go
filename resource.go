package ddd

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Scope represents the lifecycle of a dependency
type Scope int

const (
	Singleton Scope = iota
	Prototype
	Request
)

type resource struct {
	factory      reflect.Value
	typ          reflect.Type
	name         string
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	instancePool sync.Map
}

func Resource(factory any, options ...any) *resource {
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		panic("constructor must be a function")
	}

	if factoryType.NumOut() == 0 || (factoryType.NumOut() == 2 && !factoryType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
		panic("constructor must return (T) or (T, error)")
	}

	typ := factoryType.Out(0)
	name, scope := processOptions(typ, options...)

	return &resource{
		factory:      reflect.ValueOf(factory),
		typ:          typ,
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

func (r *resource) Type() reflect.Type {
	return r.typ
}

func (r *resource) TypeName() string {
	t := r.typ

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// If it has a name, use it
	if name := t.Name(); name != "" {
		return name
	}

	// Fallback to string representation
	return t.String()
}

func (r *resource) Scope() Scope {
	return r.scope
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
