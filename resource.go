package ddd

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unicode"
)

type Scope int

const (
	Singleton Scope = iota
	Prototype
)

func (s Scope) String() string {
	switch s {
	case Singleton:
		return "Singleton"
	case Prototype:
		return "Prototype"
	default:
		return "Unknown"
	}
}

var stereotypes = []reflect.Type{
	reflect.TypeOf((*Endpoint)(nil)).Elem(),
	reflect.TypeOf((*EventHandler)(nil)).Elem(),
	reflect.TypeOf((*MessageConsumer)(nil)).Elem(),
}

type resource struct {
	factory      reflect.Value
	types        []reflect.Type
	alias        string
	scope        Scope
	instance     atomic.Value
	initOnce     sync.Once
	instancePool sync.Map
}

func Resource(factory any, options ...any) *resource {

	// factory must be a function
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		panic("factory must be a function")
	}

	alias, scope := processOptions(options...)

	resource := &resource{
		factory:      reflect.ValueOf(factory),
		types:        make([]reflect.Type, 0),
		alias:        alias,
		scope:        scope,
		instancePool: sync.Map{},
	}

	resource.collectTypes()

	return resource
}

func (r *resource) Factory() any {
	return r.factory
}

// TODO change to alias
func (r *resource) Name() string {
	return r.alias
}

// Types returns all types (concrete + stereotype interfaces)
func (r *resource) Types() []reflect.Type {
	return r.types
}

func (r *resource) Scope() Scope {
	return r.scope
}

// ExecuteLifecycleHook discovers and executes a specific lifecycle hook on an instance
func ExecuteLifecycleHook(instance any, methodName string) error {
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

func (r *resource) returnType() reflect.Type {
	factoryType := r.factory.Type()

	if factoryType.NumOut() == 0 || factoryType.NumOut() > 2 {
		panic("factory must return (T) or (T, error)")
	}
	if factoryType.NumOut() == 2 {
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if !factoryType.Out(1).Implements(errorType) {
			panic("factory must return (T) or (T, error)")
		}
	}
	returnType := factoryType.Out(0)

	// // If it's a pointer, also check the element type
	// if returnType.Kind() == reflect.Ptr {
	// 	return returnType.Elem()
	// }

	return returnType
}

func (r *resource) collectTypes() {
	returnType := r.returnType()
	// Return type may be
	// a stereotype interface, we need an alias in this case to resolve
	// an arbitrary interface, alias not manadatory if there is only on such implementation, if there is no alias then we can use the camel case of interface name
	// a struct (pointer or not) that either implements an arbitrary interface and/or a stereotype interface
	r.types = append(r.types, returnType)

	if r.alias == "" {
		r.alias = ResourceName(returnType)
	}

	for _, stereotype := range stereotypes {
		if returnType.AssignableTo(stereotype) {
			// Check if not already in r.types then append
			found := false
			for _, existingType := range r.types {
				if existingType == stereotype {
					found = true
					break
				}
			}
			if !found {
				r.types = append(r.types, stereotype)
			}
		}
	}
}

func processOptions(options ...any) (string, Scope) {
	var alias string
	scope := Singleton

	for _, option := range options {
		switch v := option.(type) {
		case string:
			alias = v
		case Scope:
			scope = v
		}
	}

	return alias, scope
}

func ResourceName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	runes := []rune(t.Name())
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
