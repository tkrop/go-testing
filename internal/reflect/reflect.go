// Package reflect contains a collection of helpful generic functions that
// support reflection. It is currently not part of the public interface and
// must be consider as highly instable.
package reflect

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/tkrop/go-testing/internal/math"
)

// Aliases for types.
type (
	// Value alias for `reflect.Value`.
	Value = reflect.Value
	// Type alias for `reflect.Type`.
	Type = reflect.Type
)

// Aliases for constant values.
const (
	// Func alias for `reflect.Func`.
	Func = reflect.Func
)

// Aliases for function values.
var (
	// TypeOf alias for `reflect.TypeOf`.
	TypeOf = reflect.TypeOf
	// ValueOf alias for `reflect.ValueOf`.
	ValueOf = reflect.ValueOf
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
	if !v.IsValid() {
		return nil
	}

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

// ArgsOf returns the arguents slice for the given values.
func ArgsOf(values ...reflect.Value) []any {
	args := make([]any, 0, len(values))
	for _, value := range values {
		args = append(args, ArgOf(value))
	}

	if len(args) == 0 {
		return nil
	}
	return args
}

// ValuesIn returns the reflection values matching the input arguments of the
// given function.
func ValuesIn(ftype reflect.Type, args ...any) []reflect.Value {
	return valuesOf(typesIn(ftype, len(args)), args...)
}

// ValuesOut returns the reflection values matching the output arguments of
// the given function.
func ValuesOut(ftype reflect.Type, args ...any) []reflect.Value {
	return valuesOf(typesOut(ftype, len(args)), args...)
}

// valuesOf returns the reflection valuesOf for the given arguments. The types must
// be provided via `typesOf` that extends a variadic type function to infinity by
// returning the variadic base element on all followup indexes.
func valuesOf(ftype func(i int) reflect.Type, args ...any) []reflect.Value {
	vs := make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		t := ftype(i)
		if arg == nil {
			vs = append(vs, reflect.New(t).Elem())
		} else {
			v := reflect.ValueOf(arg)
			if !v.Type().AssignableTo(t) {
				panic(ErrInvalidType(i, t, v.Type()))
			}
			vs = append(vs, reflect.ValueOf(arg))
		}
	}

	if len(vs) == 0 {
		return nil
	}
	return vs
}

// ErrInvalidType creates a new error reporting an invalid type during value
// slice creation.
func ErrInvalidType(index int, expect, actual reflect.Type) error {
	return fmt.Errorf("invalid type at %d: expect %v got %v",
		index, expect, actual)
}

// typesOf checks the arguments length and provides the matching input type
// function from the function type. The type function is a wrapper that returns
// a variadic base type inifinitely.
func typesIn(ftype reflect.Type, args int) func(int) reflect.Type {
	num := ftype.NumIn()
	variadic := ftype.IsVariadic()
	if variadic {
		num--
	}

	if args < num {
		panic("not enough arguments")
	} else if !variadic {
		if args > num {
			panic("too many arguments")
		}
		return ftype.In
	}

	t := ftype.In(num).Elem()
	return func(i int) reflect.Type {
		if i < num {
			return ftype.In(i)
		}
		return t
	}
}

// typesOut the arguments length and provides an output type function.
func typesOut(ftype reflect.Type, args int) func(int) reflect.Type {
	num := ftype.NumOut()
	if args < num {
		panic("not enough arguments")
	} else if args > num {
		panic("too many arguments")
	}
	return ftype.Out
}

// AnyFuncOf returns a function with given number of arguments accepting any
// type.
func AnyFuncOf(args int, variadic bool) reflect.Type {
	ite := reflect.TypeOf((*any)(nil)).Elem()

	it := make([]reflect.Type, 0, args)
	for i := 0; i < args; i++ {
		if i == args-1 && variadic {
			it = append(it, reflect.TypeOf([]any{}))
		} else {
			it = append(it, ite)
		}
	}

	return reflect.FuncOf(it, []reflect.Type{}, variadic)
}

// BaseFuncOf allows to extract a base function from an interface function
// containing no receiver and suppressing the return values - if necessary. The
// given `in` and `out` values allow to restrict the included input and output
// arguments. Use `1` to remove the first argument, or `NumIn/NumOut` to remove
// all arguments.
func BaseFuncOf(mtype reflect.Type, in, out int) reflect.Type {
	it := make([]reflect.Type, 0, math.Max(mtype.NumIn()-in, 0))
	for i := in; i < mtype.NumIn(); i++ {
		it = append(it, mtype.In(i))
	}

	ot := make([]reflect.Type, 0, math.Max(mtype.NumOut()-out, 0))
	for i := out; i < mtype.NumOut(); i++ {
		ot = append(ot, mtype.Out(i))
	}

	return reflect.FuncOf(it, ot, mtype.IsVariadic())
}

// MakeFuncOf returns a newly created function with given reflective function.
func MakeFuncOf(
	mtype reflect.Type, call func([]reflect.Value) []reflect.Value,
) any {
	mvalue := reflect.New(mtype).Elem()
	mvalue.Set(reflect.MakeFunc(mtype, call))

	return mvalue.Interface()
}
