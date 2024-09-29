package test

import (
	"reflect"
	"unsafe"
)

// Builder is a generic interface that allows you to access and modify
// unexported fields of a (pointer) struct by field name.
type Builder[T any] interface {
	// Set sets the value of the field with the given name. If the name is empty,
	// and of the same type the stored target instance is replaced by the given
	// value.
	Set(name string, value any) Builder[T]
	// Get returns the value of the field with the given name. If the name is
	// empty, the stored target instance is returned.
	Get(name string) any
	// Build returns the created/modified target instance of the builder.
	Build() T
}

// Builder allows you to access and modify unexported fields of a struct.
type builder[T any] struct {
	target  any
	wrapped bool
}

// NewBuilder creates a generic builder for a target struct type. The builder
// allows you to access and modify unexported fields of the struct by field
// name.
func NewBuilder[T any]() Builder[T] {
	var target T
	return NewAccessor[T](target)
}

// NewAccessor creates a generic builder/accessor for a given target struct.
// The builder allows you to access and modify unexported fields of the struct
// by field name.
//
// If the target is a pointer to a struct (template), the pointer is stored
// and the instance is modified directly. If the target is a struct, it is
// ignored and a new pointer struct is created for modification, since a struct
// cannot be modified directly by reflection.
func NewAccessor[T any](target T) Builder[T] {
	value := reflect.ValueOf(target)

	if value.Kind() == reflect.Ptr {
		// Create a new instance if the pointer is nil.
		if value.Elem().Kind() == reflect.Invalid {
			target = reflect.New(value.Type().Elem()).Interface().(T)
			value = reflect.ValueOf(target)
		}

		if value.Elem().Kind() == reflect.Struct {
			return &builder[T]{
				target: target,
			}
		}
	} else if value.Kind() == reflect.Struct {
		// Create a new pointer instance for modification.
		value = reflect.New(value.Type())
		return &builder[T]{
			target:  value.Interface(),
			wrapped: true,
		}
	}

	panic("target must be a struct or pointer to struct")
}

// Set sets the value of the field with the given name. If the name is empty,
// and of the same type the stored target instance is replaced by the given
// value.
func (b *builder[T]) Set(name string, value any) Builder[T] {
	if name != "" {
		field := reflect.ValueOf(b.target).Elem().FieldByName(name)
		// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
		reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(value))
	} else if reflect.TypeOf(b.target) == reflect.TypeOf(value) {
		b.target = value
	} else {
		panic("target must be a compatible struct pointer")
	}
	return b
}

// Get returns the value of the field with the given name. If the name is
// empty, the stored target instance is returned.
func (b *builder[T]) Get(name string) any {
	if name != "" {
		field := reflect.ValueOf(b.target).Elem().FieldByName(name)
		// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
		return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().Interface()
	} else if b.wrapped {
		return reflect.ValueOf(b.target).Elem().Interface()
	}
	return b.target
}

// Build returns the created/modified target instance of the builder/accessor.
func (b *builder[T]) Build() T {
	if b.wrapped {
		return reflect.ValueOf(b.target).Elem().Interface().(T)
	} else {
		return b.target.(T)
	}
}
