package ddd

import (
	"fmt"
	"reflect"
	"strings"
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
	IsPointer    bool
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
	value        any // zero-initialized struct value
	resourceType string
	dependencies []Dependency
	hooks        *LifecycleHooks[any]
	scope        Scope
	initOnce     sync.Once
	instance     atomic.Value
	instancePool sync.Map
}

// Create a resource from struct to be added to the context to be autowired
// and instatiated as a singelton, prototype or request scoped resource instance
func NewResource[I any](options ...any) *Resource {

	var valueType reflect.Type
	value, name, scope := processOptions(options...)

	// get generic type
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	if interfaceType.Kind() == reflect.Interface {
		if value == nil {
			panic("options must include a struct if resource generic type is an interface")
		}
		valueType = reflect.TypeOf(value)

		if !valueType.Implements(interfaceType) {
			panic(fmt.Sprintf("value of type %v does not implement interface %v", valueType, interfaceType))
		}

		// } else {
		// 	panic("value must be a struct type or pointer to struct")
		// }

	} else if interfaceType.Kind() == reflect.Struct {
		value = *new(I) // non-pointer zero struct
		valueType = interfaceType
	} else {
		panic("type parameter must be an interface or struct")
	}

	if name == "" {
		name = toCamelCase(valueType.Name())
	}
	// Make sure value is a struct kind type or pointer to struct
	dependencies := parseDependencies(valueType)

	lifecycleHooks := getLifecycleHooks(value, valueType)

	return &Resource{
		name:         name,
		resourceType: getSimpleTypeName(interfaceType.String()),
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

func (r *Resource) Value() any {
	return r.value
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

func processOptions(options ...any) (any, string, Scope) {
	var name string
	var value any
	scope := Singleton

	// Process each option
	for _, option := range options {
		switch v := option.(type) {
		case string:
			name = v
		case Scope:
			scope = v
		default:
			valueType := reflect.TypeOf(option)
			// if valueType.Kind() == reflect.Ptr {
			// 	valueType = valueType.Elem()
			// }
			if valueType.Kind() != reflect.Struct {
				panic("value must be a struct type")
			}
			value = option

			if name == "" {
				name = getDefaultName(valueType)
			}
		}
	}

	return value, name, scope
}

func parseDependencies(valueType reflect.Type) []Dependency {
	var dependencies []Dependency
	for i := range valueType.NumField() {
		field := valueType.Field(i)

		// Check if the field has a resource tag
		_, hasTag := field.Tag.Lookup(ResourceTag)

		// Only consider fields with a resource tag
		if hasTag {
			// Skip primitive types regardless of whether they have a resource tag
			if isPrimitiveType(field.Type) {
				continue
			}

			// Get the tag value
			tagValue := field.Tag.Get(ResourceTag)

			// Determine the resource name
			resourceName := toCamelCase(field.Name)
			if tagValue != "" { // If tag has a value, use that as the resource name
				resourceName = tagValue
			}

			// Determine if the field is a pointer
			isPointer := field.Type.Kind() == reflect.Ptr

			// Use the actual type of the field as ResourceType
			dependencies = append(dependencies, Dependency{
				ResourceType: field.Type.String(),
				ResourceName: resourceName,
				IsPointer:    isPointer,
			})
		}
	}
	return dependencies
}

func getDefaultName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toCamelCase(t.Name())
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func getSimpleTypeName(fullName string) string {
	// Check if we have a generic type
	openBracketIndex := strings.Index(fullName, "[")
	if openBracketIndex == -1 {
		return fullName
	}

	closeBracketIndex := strings.LastIndex(fullName, "]")
	if closeBracketIndex == -1 {
		return fullName
	}

	// Get the base type
	baseType := fullName[:openBracketIndex]

	// Extract the type parameter(s)
	fullTypeParamSection := fullName[openBracketIndex+1 : closeBracketIndex]

	// Handle multiple type parameters
	typeParams := strings.Split(fullTypeParamSection, ",")
	simplifiedParams := make([]string, 0, len(typeParams))

	for _, param := range typeParams {
		param = strings.TrimSpace(param)

		// Get just the type name after the last dot
		lastDotIndex := strings.LastIndex(param, ".")
		if lastDotIndex == -1 {
			simplifiedParams = append(simplifiedParams, param) // No dot found, use as is
		} else {
			simplifiedParams = append(simplifiedParams, param[lastDotIndex+1:])
		}
	}

	// Reconstruct the simplified type
	return baseType + "[" + strings.Join(simplifiedParams, ", ") + "]"
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
