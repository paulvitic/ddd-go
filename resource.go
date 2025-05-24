package ddd

import (
	"reflect"
)

// Scope represents the lifecycle of a dependency
type Scope int

const (
	Singleton Scope = iota
	Prototype
	Request
)

type resource struct {
	value        any
	resourceType reflect.Type
	name         string
	scope        Scope
	hooks        any
}

func Resource(constructor any, options ...any) *resource {
	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		panic("constructor must be a function")
	}

	if constructorType.NumOut() == 0 || (constructorType.NumOut() == 2 && !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem())) {
		panic("constructor must return (T) or (T, error)")
	}

	typ := constructorType.Out(0)
	name, scope, hooks := processOptions(typ, options...)

	return &resource{
		value:        constructor,
		resourceType: typ,
		name:         name,
		scope:        scope,
		hooks:        hooks,
	}
}

func (r *resource) Value() any {
	return r.value
}

func (r *resource) Name() string {
	return r.name
}

func (r *resource) Type() reflect.Type {
	return r.resourceType
}

func (r *resource) Scope() Scope {
	return r.scope
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
