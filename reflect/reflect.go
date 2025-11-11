package reflect

import (
	"reflect"
	"slices"
	"strings"
	"unsafe"
)

// Getter is a generic interface that allows you to access unexported fields
// of a (pointer) struct by field name.
type Getter[T any] interface {
	// Get returns the value of the field with the given name. If the name is
	// empty, the stored target instance is returned.
	Get(name string) any
}

// Setter is a generic fluent interface that allows you to modify unexported
// fields of a (pointer) struct by field name.
type Setter[T any] interface {
	// Set sets the value of the field with the given name. If the name is empty,
	// and of the same type the stored target instance is replaced by the given
	// value.
	Set(name string, value any) Setter[T]
	// Build returns the created or modified target instance of the builder.
	Build() T
}

// Finder is a generic interface that allows you to access unexported fields
// of a (pointer) struct by field name.
type Finder[T any] interface {
	// Find returns the first value of a field from the given list of field
	// names with a type matching the default value type. If the name list is
	// empty or contains a star (`*`), the first matching field in order of the
	// struct declaration is returned as fallback. If no matching field is
	// found, the default value is returned.
	Find(dflt any, names ...string) any
}

// Builder is a generic, partially fluent interface that allows you to access
// and modify unexported fields of a (pointer) struct by field name.
type Builder[T any] interface {
	// Getter is a generic interface that allows you to access unexported fields
	// of a (pointer) struct by field name.
	Getter[T]
	// Finder is a generic interface that allows you to access unexported fields
	// of a (pointer) struct by field name.
	Finder[T]
	// Setter is a generic fluent interface that allows you to modify unexported
	// fields of a (pointer) struct by field name.
	Setter[T]
}

// Builder is used for accessing and modifying unexported fields in a struct
// or a struct pointer.
type builder[T any] struct {
	// target is the struct reflection value of the struct or struct pointer
	// instance to be accessed and modified.
	target any
	// rtype is the targets reflection type of the struct or struct pointer
	// instance to be accessed and modified.
	rtype reflect.Type
	// wrapped is true if the target is a struct that is actually wrapped in
	// a pointer instance.
	wrapped bool
}

// NewBuilder creates a generic builder for a target struct type. The builder
// allows you to access and modify unexported fields of the struct by field
// name.
func NewBuilder[T any]() Builder[T] {
	var target T
	return NewAccessor[T](target)
}

// NewGetter creates a generic getter for a target struct type. The getter
// allows you to access unexported fields of the struct by field name.
func NewGetter[T any](target T) Getter[T] {
	return NewAccessor[T](target)
}

// NewSetter creates a generic setter for a target struct type. The setter
// allows you to modify unexported fields of the struct by field name.
func NewSetter[T any](target T) Setter[T] {
	return NewAccessor[T](target)
}

// NewFinder creates a generic finder for a target struct type. The finder
// allows you to access unexported fields of the struct by field name.
func NewFinder[T any](target T) Finder[T] {
	return NewAccessor[T](target)
}

// NewAccessor creates a generic builder/accessor for a given target struct.
// The builder allows you to access and modify unexported fields of the struct
// by field name.
//
// If the target is a pointer to a struct (template), the pointer is stored
// and the instance is modified directly. If the pointer is nil a new instance
// is created and stored for modification.
//
// If the target is a struct, it cannot be modified directly and a new pointer
// struct is created to circumvent the access restrictions on private fields.
// The pointer struct is stored for modification.
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
				target:  target,
				rtype:   value.Elem().Type(),
				wrapped: false,
			}
		}
	} else if value.Kind() == reflect.Struct {
		// Create a new pointer instance for modification.
		value = reflect.New(value.Type())
		value.Elem().Set(reflect.ValueOf(target))
		return &builder[T]{
			target:  value.Interface(),
			rtype:   value.Elem().Type(),
			wrapped: true,
		}
	}

	return nil
}

// Set sets the value of the field with the given name. If the name is empty,
// and of the same type the stored target instance is replaced by the given
// value. If the value is nil, the field is set to the zero value of the field
// type. If the field is not found or the value is not assignable to it, a
// panic is raised.
func (b *builder[T]) Set(name string, value any) Setter[T] {
	if name != "" {
		field := b.fieldValueOf(name, value)
		b.valuePtr(field).Elem().Set(b.valueOf(field, value))
	} else if value == nil && b.rtype.Kind() == reflect.Struct {
		b.target = reflect.New(b.rtype).Interface()
	} else if reflect.TypeOf(b.target) == reflect.TypeOf(value) {
		b.target = value
	} else {
		panic("target must be compatible struct pointer [" +
			reflect.TypeOf(value).String() + " => " +
			reflect.TypeOf(b.target).String() + "]")
	}
	return b
}

// Get returns the value of the field with the given name. If the name is
// empty, the stored target instance is returned.
func (b *builder[T]) Get(name string) any {
	if name == "" {
		return b.Build()
	}

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
		for i := range b.rtype.NumField() {
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

// fieldValueOf returns the reflect value of the field with the given name.
// If the field is not found or the value is not assignable to found field, a
// panic is raised. The method is used to ensure that the field exists and
// that the value is compatible with the field type before setting it.
func (b *builder[T]) fieldValueOf(name string, value any) reflect.Value {
	field := b.targetValueOf().FieldByName(name)
	if !field.IsValid() {
		panic("target field not found [" + name + "]")
	} else if !b.canBeAssigned(field.Type(), reflect.TypeOf(value)) {
		panic("value must be compatible [" + reflect.TypeOf(value).
			String() + " => " + field.Type().String() + "]")
	}
	return field
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
		return value.AssignableTo(field)
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

// Find returns the first value of a struct field from the given list of field
// names with a type matching the default value type. If the name list is empty
// or contains a star (`*`), the first matching field in order of the struct
// declaration is returned as fallback. If no matching field is found, the
// default value is returned.
//
// The `param` object can be a struct, a pointer to a struct, or an arbitrary
// value matching the default value type. In the last case, the arbitrary value
// is returned as is.
func Find[P, T any](param P, deflt T, names ...string) T {
	pt, dt := reflect.TypeOf(param), reflect.TypeOf(deflt)
	switch {
	case pt.Kind() == dt.Kind():
		return reflect.ValueOf(param).Interface().(T)
	case pt.Kind() == reflect.Struct:
		return NewAccessor[P](param).Find(deflt, names...).(T)
	case pt.Kind() == reflect.Ptr && pt.Elem().Kind() == reflect.Struct:
		return NewAccessor[P](param).Find(deflt, names...).(T)
	default:
		return deflt
	}
}

// Name returns the normalized test case name for the given default name and
// parameter set. If the default name is empty, the test name is resolved from
// the parameter set using the `name` field. The resolved value is normalized
// before being returned. If no test name can be resolved an empty string is
// returned.
func Name[P any](name string, param P) string {
	if name != "" {
		return strings.ReplaceAll(name, " ", "-")
	} else if name := Find(param, "", "name", "Name"); name != "" {
		return strings.ReplaceAll(name, " ", "-")
	}
	return ""
}
