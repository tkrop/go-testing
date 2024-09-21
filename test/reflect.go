package test

import (
	"reflect"
	"unsafe"
)

// Error creates an Accessor for the given error to access and modify its
// unexported fields by field name.
//
// Example:
//
//	err := test.Error(errors.New("error message")).Set("text", "new message").Get("")
//	fmt.Println(err.Error()) // Output: new message
//
//	err := test.Error(errors.New("error message")).Set("text", "new message").Get("text")
//	fmt.Println(err) // Output: new message
func Error(err error) *Accessor[error] {
	return NewAccessor[error](err)
}

// Accessor allows you to access and modify unexported fields of a struct.
type Accessor[T any] struct {
	target T
}

// NewAccessor creates a generic accessor for the given target.
func NewAccessor[T any](target T) *Accessor[T] {
	return &Accessor[T]{target: target}
}

// Set sets the value of the accessor target's field with the given name.
func (a *Accessor[T]) Set(name string, value any) *Accessor[T] {
	field := reflect.ValueOf(a.target).Elem().FieldByName(name)
	// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().Set(reflect.ValueOf(value))

	return a
}

// Get returns the value of the field with the given name. If the name is empty,
// it returns the accessor target itself.
func (a *Accessor[T]) Get(name string) any {
	if name == "" {
		return a.target
	}
	field := reflect.ValueOf(a.target).Elem().FieldByName(name)
	// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().Interface()
}
