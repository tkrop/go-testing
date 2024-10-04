package test

import (
	"reflect"
	"slices"
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
	// Find returns the first value of a field from the given list of field
	// names with a type matching the default value type. If the name list is
	// empty or contains a star (`*`), the first matching field in order of the
	// struct declaration is returned as fallback. If no matching field is
	// found, the default value is returned.
	Find(dflt any, names ...string) any
	// Build returns the created/modified target instance of the builder.
	Build() T
}

// Find returns the first value of a parameter field from the given list of
// field names with a type matching the default value type. If the name list
// is empty or contains a star (`*`), the first matching field in order of the
// struct declaration is returned as fallback. If no matching field is found,
// the default value is returned.
//
// The `param“ object can be a struct, a pointer to a struct, or an arbitrary
// value matching the default value type. In the last case, the arbitrary value
// is returned as is.
func Find[P, T any](param P, deflt T, names ...string) T {
	pt, dt := reflect.TypeOf(param), reflect.TypeOf(deflt)
	if pt.Kind() == dt.Kind() {
		return reflect.ValueOf(param).Interface().(T)
	} else if pt.Kind() == reflect.Struct {
		// TODO: This is currently not working as expected and creates panics.
		return NewAccessor[*P](&param).Find(deflt, names...).(T)
	} else if pt.Kind() == reflect.Ptr && pt.Elem().Kind() == reflect.Struct {
		return NewAccessor[P](param).Find(deflt, names...).(T)
	}
	return deflt
}

// Builder allows you to access and modify unexported fields of a struct.
type builder[T any] struct {
	target  any
	rtype   reflect.Type
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
				rtype:  value.Elem().Type(),
			}
		}
	} else if value.Kind() == reflect.Struct {
		// Create a new pointer instance for modification.
		value = reflect.New(value.Type())
		return &builder[T]{
			target:  value.Interface(),
			rtype:   value.Elem().Type(),
			wrapped: true,
		}
	}

	panic("target must be struct or struct pointer [" +
		typeOf(target) + "]")
}

// typeOf returns the type of the given target instance. If the target is nil,
// the type is "nil".
func typeOf(target any) string {
	if target != nil {
		return reflect.TypeOf(target).String()
	}
	return "nil"
}

// Set sets the value of the field with the given name. If the name is empty,
// and of the same type the stored target instance is replaced by the given
// value. If the value is nil, the field is set to the zero value of the field
// type. If the field is not found or the value is not assignable to it, a
// panic is raised.
func (b *builder[T]) Set(name string, value any) Builder[T] {
	if name != "" {
		b.set(name, value)
	} else if value == nil && b.rtype.Kind() == reflect.Struct {
		b.target = reflect.New(b.rtype).Interface()
	} else if reflect.TypeOf(b.target) == reflect.TypeOf(value) {
		b.target = value
	} else {
		panic("target must be compatible struct pointer [" +
			typeOf(value) + " => " + typeOf(b.target) + "]")
	}
	return b
}

// Get returns the value of the field with the given name. If the name is
// empty, the stored target instance is returned.
func (b *builder[T]) Get(name string) any {
	if name != "" {
		target := b.targetValueOf()
		if !target.IsValid() {
			if field, ok := b.rtype.FieldByName(name); ok {
				return reflect.New(field.Type).Elem().Interface()
			}
			panic("target field not found [" + name + "]")
		}
		field := target.FieldByName(name)
		if !field.IsValid() {
			panic("target field not found [" + name + "]")
		}
		return b.valuePtr(field).Elem().Interface()
	}
	return b.Build()
}

// Find returns the first value of a field from the given list of field names
// with a type matching the default value type. If the name list is empty or
// contains a star (`*`), the first matching field in order of the struct
// declaration is returned as fallback. If no matching field is found, the
// default value is returned.
func (b *builder[T]) Find(deflt any, names ...string) any {
	for _, name := range names {
		if field := b.targetValueOf().FieldByName(name); field.IsValid() {
			if b.canBeAssigned(reflect.TypeOf(deflt), field.Type()) {
				return b.valuePtr(field).Elem().Interface()
			}
		}
	}

	if len(names) == 0 || slices.Contains(names, "*") {
		// Fallback to the first field with a matching type.
		for i := 0; i < b.rtype.NumField(); i++ {
			tfield := b.rtype.Field(i)
			if b.canBeAssigned(reflect.TypeOf(deflt), tfield.Type) {
				vfield := b.targetValueOf().Field(i)
				return b.valuePtr(vfield).Elem().Interface()
			}
		}
	}
	return deflt
}

// Build returns the created/modified target instance of the builder/accessor.
func (b *builder[T]) Build() T {
	if b.wrapped {
		target := b.targetValueOf()
		if target.IsValid() {
			return target.Interface().(T)
		}
		var t T
		return t
	} else {
		return b.target.(T)
	}
}

// set sets the value of the field with the given name. If the value is nil,
// the field is set to the zero value of the field type. If the field is not
// found or the value is not assignable to it, a panic is raised.
func (b *builder[T]) set(name string, value any) {
	field := b.targetValueOf().FieldByName(name)
	if !field.IsValid() {
		panic("target field not found [" + name + "]")
	} else if !b.canBeAssigned(field.Type(), reflect.TypeOf(value)) {
		panic("value must be compatible [" +
			typeOf(value) + " => " + field.Type().String() + "]")
	}
	b.valuePtr(field).Elem().Set(b.valueOf(field, value))
}

// targetValueOf returns the reflect value of the target instance.
func (b *builder[T]) targetValueOf() reflect.Value {
	return reflect.ValueOf(b.target).Elem()
}

// canBeAssigned returns true if the given value is compatible with the given
// field type. This is either the case if the value implements the field type
// or if the value is assignable to the field type. Nil value types are always
// assignable using the zero value.
func (builder[T]) canBeAssigned(field reflect.Type, value reflect.Type) bool {
	if value == nil {
		return true
	} else if field.Kind() == reflect.Interface {
		return value.Implements(field)
	} else {
		return field.AssignableTo(value)
	}
}

// valuePtr returns the reflect value of the field as a pointer. This is
// required to set the value of a field, since reflect.Value.Set does not
// support setting unexported fields.
func (*builder[T]) valuePtr(field reflect.Value) reflect.Value {
	// #nosec G103,G115 // This is a safe use of unsafe.Pointer.
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
}

// valueOf returns the reflect value of the given value. If the value is nil,
// the zero value of the field type is returned.
func (*builder[T]) valueOf(field reflect.Value, value any) reflect.Value {
	if value == nil {
		return reflect.Zero(field.Type())
	}
	return reflect.ValueOf(value)
}
