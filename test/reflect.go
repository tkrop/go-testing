package test

import (
	"reflect"
	"unsafe"
)

// Error creates an accessor/build for the given error to access and modify its
// unexported fields by field name.
func Error(err error) *Accessor[error] {
	return NewAccessor[error](err)
}

// Accessor allows you to access and modify unexported fields of a struct.
type Accessor[T any] struct {
	target  T
	wrapped bool
}

// NewAccessor creates a generic accessor/builder for a given target struct.
// If the target is a pointer to a struct (template), the instance is stored
// and modified. If the target is a struct, a pointer to a new instance of is
// created, since a struct cannot be modified by reflection.
func NewAccessor[T any](target T) *Accessor[T] {
	value := reflect.ValueOf(target)
	if value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
		return &Accessor[T]{
			target: value.Interface().(T),
		}
	} else if value.Kind() == reflect.Struct {
		target = reflect.New(value.Type()).Interface().(T)

		return &Accessor[T]{
			target:  target,
			wrapped: true,
		}
	}
	panic("target must be a struct or pointer to struct")
}

// Set sets the value of the field with the given name. If the name is empty,
// and of the same type the stored target instance is replaced by the given
// value.
func (a *Accessor[T]) Set(name string, value any) *Accessor[T] {
	if name != "" {
		field := reflect.ValueOf(a.target).Elem().FieldByName(name)
		// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
		reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(value))
	} else if reflect.TypeOf(a.target) == reflect.TypeOf(value) {
		a.target = value.(T)
	} else {
		panic("target must of compatible struct pointer type")
	}
	return a
}

// Get returns the value of the field with the given name. If the name is
// empty, the stored target instance is returned.
func (a *Accessor[T]) Get(name string) any {
	if name != "" {
		field := reflect.ValueOf(a.target).Elem().FieldByName(name)
		// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
		return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().Interface()
	} else if a.wrapped {
		return reflect.ValueOf(a.target).Elem().Interface()
	}
	return a.target
}
