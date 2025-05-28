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
	hooks        LifecycleHooks[any]
}

// LifecycleHooks defines lifecycle hooks for dependencies
type LifecycleHooks[T any] struct {
	OnInit    func(T) error
	OnStart   func(T) error
	OnDestroy func(T) error
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
	name, scope, hooks := processOptions(typ, options...)

	return &resource{
		factory:      reflect.ValueOf(factory),
		typ:          typ,
		name:         name,
		scope:        scope,
		hooks:        hooks,
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

func processOptions(typ reflect.Type, options ...any) (string, Scope, LifecycleHooks[any]) {
	var name string
	scope := Singleton
	var hooks LifecycleHooks[any]

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

func isLifecycleHooks(v any) (LifecycleHooks[any], bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return LifecycleHooks[any]{}, false
	}

	rt := rv.Type()
	if rt.NumField() != 3 {
		return LifecycleHooks[any]{}, false
	}

	onInitField, hasOnInit := rt.FieldByName("OnInit")
	onStartField, hasOnStart := rt.FieldByName("OnStart")
	onDestroyField, hasOnDestroy := rt.FieldByName("OnDestroy")

	if !hasOnInit || !hasOnStart || !hasOnDestroy {
		return LifecycleHooks[any]{}, false
	}

	isValidHook := func(f reflect.StructField) bool {
		return f.Type.Kind() == reflect.Func &&
			f.Type.NumIn() == 1 &&
			f.Type.NumOut() == 1 &&
			f.Type.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
	}

	if !isValidHook(onInitField) || !isValidHook(onStartField) || !isValidHook(onDestroyField) {
		return LifecycleHooks[any]{}, false
	}

	return LifecycleHooks[any]{
		OnInit:    convertToInterfaceFunc(rv.FieldByName("OnInit")),
		OnStart:   convertToInterfaceFunc(rv.FieldByName("OnStart")),
		OnDestroy: convertToInterfaceFunc(rv.FieldByName("OnDestroy")),
	}, true
}
