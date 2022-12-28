// Package reflect contains a collection of helpful generic functions helping
// with reflection. It is currently not part of the public interface and must
// be consider as highly instable.
package reflect

import (
	"reflect"
	"unsafe"
)

// Aliases for types.
type (
	// Value alias for `reflect.Value`.
	Value = reflect.Value
	// Type alias for `reflect.Type`.
	Type = reflect.Type
)

// Aliases for functions and values.
var (
	// TypeOf alias for `reflect.TypeOf`.
	TypeOf = reflect.TypeOf
	// ValueOf alias for `reflect.ValueOf`.
	ValueOf = reflect.ValueOf
	// Func alias for `reflect.Func`.
	Func = reflect.Func
)

// FindArgOf find the first argument with one of the given field names matching
// the type matching the default argument type.
func FindArgOf[P any](param P, deflt any, names ...string) any {
	t := reflect.TypeOf(param)
	dt := reflect.TypeOf(deflt)
	if t.Kind() != reflect.Struct {
		if t.Kind() == dt.Kind() {
			return param
		}
		return deflt
	}

	v := reflect.ValueOf(param)

	found := false
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		if fv.Type().Kind() == dt.Kind() {
			for _, name := range names {
				if t.Field(i).Name == name {
					return FieldArgOf(v, i)
				}
			}
			if !found {
				deflt = FieldArgOf(v, i)
				found = true
			}
		}
	}
	return deflt
}

// FieldArgOf returns the argument of the `i`th field of the given value.
func FieldArgOf(v reflect.Value, i int) any {
	vf := v.Field(i)
	if vf.CanInterface() {
		return ArgOf(vf)
	}

	// Make a copy to circumvent access restrictions.
	vr := reflect.New(v.Type()).Elem()
	vr.Set(v)

	// Get the field value from the copy.
	vf = vr.Field(i)
	rf := reflect.NewAt(vf.Type(), unsafe.Pointer(vf.UnsafeAddr())).Elem()

	var value any
	reflect.ValueOf(&value).Elem().Set(rf)
	return value
}

// ArgOf returns the argument of the given value.
func ArgOf(v reflect.Value) any {
	switch v.Type().Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int:
		return int(v.Int())
	case reflect.Int8:
		return int8(v.Int())
	case reflect.Int16:
		return int16(v.Int())
	case reflect.Int32:
		return int32(v.Int())
	case reflect.Int64:
		return v.Int()
	case reflect.Uint:
		return uint(v.Uint())
	case reflect.Uint8:
		return uint8(v.Uint())
	case reflect.Uint16:
		return uint16(v.Uint())
	case reflect.Uint32:
		return uint32(v.Uint())
	case reflect.Uint64:
		return v.Uint()
	// TODO find test case.
	// case reflect.Uintptr:
	// 	return v.Pointer()
	case reflect.Float32:
		return float32(v.Float())
	case reflect.Float64:
		return v.Float()
	case reflect.Complex64:
		return complex64(v.Complex())
	case reflect.Complex128:
		return v.Complex()
	case reflect.String:
		return v.String()
	default:
		return v.Interface()
	}
}
