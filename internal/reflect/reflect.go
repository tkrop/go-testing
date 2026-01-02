package reflect

import (
	"errors"
	"fmt"
	"reflect"
)

// Aliases for types.
type (
	// Value alias for `reflect.Value`.
	Value = reflect.Value
	// Type alias for `reflect.Type`.
	Type = reflect.Type
	// Kind alias for `reflect.Kind`.
	Kind = reflect.Kind
)

// Aliases for constant values.
const (
	// Invalid alias for `reflect.Invalid`.
	Invalid = reflect.Invalid
	// Bool alias for `reflect.Bool`.
	Bool = reflect.Bool
	// Int alias for `reflect.Int`.
	Int = reflect.Int
	// Int8 alias for `reflect.Int8`.
	Int8 = reflect.Int8
	// Int16 alias for `reflect.Int16`.
	Int16 = reflect.Int16
	// Int32 alias for `reflect.Int32`.
	Int32 = reflect.Int32
	// Int64 alias for `reflect.Int64`.
	Int64 = reflect.Int64
	// Uint alias for `reflect.Uint`.
	Uint = reflect.Uint
	// Uint8 alias for `reflect.Uint8`.
	Uint8 = reflect.Uint8
	// Uint16 alias for `reflect.Uint16`.
	Uint16 = reflect.Uint16
	// Uint32 alias for `reflect.Uint32`.
	Uint32 = reflect.Uint32
	// Uint64 alias for `reflect.Uint64`.
	Uint64 = reflect.Uint64
	// Uintptr alias for `reflect.Uintptr`.
	Uintptr = reflect.Uintptr
	// Float32 alias for `reflect.Float32`.
	Float32 = reflect.Float32
	// Float64 alias for `reflect.Float64`.
	Float64 = reflect.Float64
	// Complex64 alias for `reflect.Complex64`.
	Complex64 = reflect.Complex64
	// Complex128 alias for `reflect.Complex128`.
	Complex128 = reflect.Complex128
	// Array alias for `reflect.Array`.
	Array = reflect.Array
	// Chan alias for `reflect.Chan`.
	Chan = reflect.Chan
	// Func alias for `reflect.Func`.
	Func = reflect.Func
	// Interface alias for `reflect.Interface`.
	Interface = reflect.Interface
	// Map alias for `reflect.Map`.
	Map = reflect.Map
	// Pointer alias for `reflect.Pointer`.
	Pointer = reflect.Pointer
	// Slice alias for `reflect.Slice`.
	Slice = reflect.Slice
	// String alias for `reflect.String`.
	String = reflect.String
	// Struct alias for `reflect.Struct`.
	Struct = reflect.Struct
	// UnsafePointer alias for `reflect.UnsafePointer`.
	UnsafePointer = reflect.UnsafePointer
)

// Aliases for function values.
var (
	// TypeOf alias for `reflect.TypeOf`.
	TypeOf = reflect.TypeOf
	// ValueOf alias for `reflect.ValueOf`.
	ValueOf = reflect.ValueOf
	// New alias for `reflect.New`.
	New = reflect.New
	// Zero alias for `reflect.Zero`.
	Zero = reflect.Zero
	// Ptr alias for `reflect.Ptr`.
	Ptr = reflect.Ptr
	// PointerTo alias for `reflect.PointerTo`.
	PointerTo = reflect.PointerTo
)

// ArgOf returns the argument of the given value.
func ArgOf(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	switch v.Type().Kind() { //nolint:exhaustive // covered by default.
	case reflect.Bool:
		return v.Bool()
	case reflect.Int:
		return int(v.Int())
	case reflect.Int8:
		return int8(v.Int()) // #nosec G115 // checked by type switch.
	case reflect.Int16:
		return int16(v.Int()) // #nosec G115 // checked by type switch.
	case reflect.Int32:
		return int32(v.Int()) // #nosec G115 // checked by type switch.
	case reflect.Int64:
		return v.Int()
	case reflect.Uint:
		return uint(v.Uint())
	case reflect.Uint8:
		return uint8(v.Uint()) // #nosec G115 // checked by type switch.
	case reflect.Uint16:
		return uint16(v.Uint()) // #nosec G115 // checked by type switch.
	case reflect.Uint32:
		return uint32(v.Uint()) // #nosec G115 // checked by type switch.
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

// ArgsOf returns the arguments slice for the given values.
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
	return valuesOf(typesIn(ftype, args...))
}

// ValuesOut returns the reflection values matching the output arguments of
// the given function. If lenient is set, non existing output arguments are
// filled with zero values to support incomplete argument lists.
func ValuesOut(ftype reflect.Type, lenient bool, args ...any) []reflect.Value {
	return valuesOf(typesOut(ftype, lenient, args...))
}

// typesIn checks the input argument length and provides the matching input
// type function, the number of arguments, and the default values. The input
// type function is a wrapper the standard input type function that returns a
// variadic base type for all arguments exceeding the function argument number.
func typesIn(
	ftype reflect.Type, args ...any,
) (func(int) reflect.Type, int, []any) {
	num := ftype.NumIn()
	variadic := ftype.IsVariadic()
	if variadic {
		num--
	}

	anum := len(args)
	if anum < num {
		panic("not enough arguments")
	} else if !variadic {
		if anum > num {
			panic("too many arguments")
		}
		return ftype.In, anum, args
	}

	t := ftype.In(num).Elem()
	return func(i int) reflect.Type {
		if i < num {
			return ftype.In(i)
		}
		return t
	}, anum, args
}

// typesOut checks the output arguments number against the function expectation
// if not lenient is give and provides an output type function, the number of
// output arguments, and the default values.
func typesOut(
	ftype reflect.Type, lenient bool, args ...any,
) (func(int) reflect.Type, int, []any) {
	num := ftype.NumOut()
	if !lenient {
		anum := len(args)
		if anum < num {
			panic("not enough arguments")
		} else if anum > num {
			panic("too many arguments")
		}
	}
	return ftype.Out, num, args
}

// valuesOf returns the reflection values for the given arguments matching the
// given reflection types. The types must be provided via `typesOf`-function
// that extends a variadic type function to infinity by returning the variadic
// base element on all followup indexes. The `num` input gives the number of
// arguments to be created from the given arguments.
func valuesOf(
	ftype func(i int) reflect.Type, num int, args []any,
) []reflect.Value {
	vs := make([]reflect.Value, 0, num)
	for i := range num {
		t := ftype(i)
		arg := argOrNil(i, args...)
		if arg == nil {
			vs = append(vs, reflect.New(t).Elem())
		} else {
			v := reflect.ValueOf(arg)
			if !v.Type().AssignableTo(t) {
				panic(NewErrInvalidType(i, t, v.Type()))
			}
			vs = append(vs, reflect.ValueOf(arg))
		}
	}

	if len(vs) == 0 {
		return nil
	}
	return vs
}

// argOrNil returns the `i`-th argument, if the index is within the boundaries,
// or `nil`, if it is out of bounds.
func argOrNil(i int, args ...any) any {
	if i < len(args) {
		return args[i]
	}
	return nil
}

var errInvalidType = errors.New("invalid type")

// NewErrInvalidType creates a new error reporting an invalid type during value
// slice creation.
func NewErrInvalidType(index int, expect, actual reflect.Type) error {
	return fmt.Errorf("%w at %d: expect %v got %v",
		errInvalidType, index, expect, actual)
}

// AnyFuncOf returns a function with given number of arguments accepting any
// type.
func AnyFuncOf(args int, variadic bool) reflect.Type {
	ite := reflect.TypeOf((*any)(nil)).Elem()

	it := make([]reflect.Type, 0, args)
	for i := range args {
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
	it := make([]reflect.Type, 0, max(mtype.NumIn()-in, 0))
	for i := in; i < mtype.NumIn(); i++ {
		it = append(it, mtype.In(i))
	}

	ot := make([]reflect.Type, 0, max(mtype.NumOut()-out, 0))
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
