package reflect

import (
	"math/rand"
	"reflect"
	"unsafe"
)

// random is an implementation of Random interface.
type random struct {
	rand   *rand.Rand
	size   int
	length int
}

// Random is an interface for generating random data structures.
type Random interface {
	Random(obj any) any
}

// NewRandom creates a random generator with default size and length limits.
func NewRandom(seed int64, size, length int) Random {
	return &random{
		// #nosec G404 -- Intentional use for testing.
		rand:   rand.New(rand.NewSource(seed)),
		size:   size,
		length: length,
	}
}

// Random generates random data into an existing data structure filling in
// gaps. If this is not possible, a new value is allocated and returned.
func (r *random) Random(obj any) any {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	k := v.Kind()

	if isPrimitiveKind(k) { // primitive by value
		return r.newPrimitive(k)
	}

	switch k {
	case reflect.Ptr:
		if v.IsNil() {
			v = reflect.New(v.Type().Elem())
			obj = v.Interface()
		}
		r.randomField(v.Elem())
		return obj
	case reflect.Slice:
		return r.randomSliceType(v.Type())
	case reflect.Map:
		return r.randomMap(v.Type())
	case reflect.Struct:
		value := reflect.New(v.Type()).Elem()
		r.randomStruct(value)
		return value.Interface()
	default:
		return obj
	}
}

// isPrimitiveKind checks if a reflect.Kind is a primitive type.
func isPrimitiveKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

// newPrimitive generates a random value for a given primitive kind.
func (r *random) newPrimitive(kind reflect.Kind) any {
	switch kind {
	case reflect.Bool:
		return r.rand.Intn(r.length) != 0
	case reflect.Int:
		return r.rand.Intn(r.length) + 1
	case reflect.Int8:
		return int8(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Int16:
		return int16(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Int32:
		return int32(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Int64:
		return int64(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uint:
		return uint(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uint8:
		return uint8(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uint16:
		return uint16(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uint32:
		return uint32(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uint64:
		return uint64(r.rand.Intn(r.length) + 1) // #nosec G115 -- intentional use.
	case reflect.Uintptr:
		return uintptr(r.rand.Intn(r.length) + 1)
	case reflect.Float32:
		return float32(r.rand.Float64())
	case reflect.Float64:
		return r.rand.Float64()
	case reflect.Complex64:
		return complex(float32(r.rand.Float64()), float32(r.rand.Float64()))
	case reflect.Complex128:
		return complex(r.rand.Float64(), r.rand.Float64())
	case reflect.String:
		return r.randomString()
	default:
		return nil
	}
}

// randomStruct fills in the fields of a struct with random data.
func (r *random) randomStruct(v reflect.Value) {
	for i := range v.NumField() {
		field := v.Field(i)
		r.randomField(field)
	}
}

// randomSliceType generates a random slice of the given type.
func (r *random) randomSliceType(t reflect.Type) any {
	ln := r.rand.Intn(r.size) + 1
	s := reflect.MakeSlice(t, ln, ln)
	for i := range ln {
		r.randomField(s.Index(i))
	}
	return s.Interface()
}

// randomMap generates a random map of the given type.
func (r *random) randomMap(t reflect.Type) any {
	m := reflect.MakeMap(t)
	ln := r.rand.Intn(r.size) + 1
	for range ln {
		k := reflect.New(t.Key()).Elem()
		v := reflect.New(t.Elem()).Elem()
		r.randomField(k)
		r.randomField(v)
		m.SetMapIndex(k, v)
	}
	return m.Interface()
}

// randomField fills in a field with random data based on its kind.
func (r *random) randomField(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			r.setField(v, reflect.New(v.Type().Elem()))
		}
		r.randomField(v.Elem())
	case reflect.Struct:
		r.randomStruct(v)
	case reflect.Slice:
		newVal := r.randomSliceType(v.Type())
		r.setField(v, reflect.ValueOf(newVal))
	case reflect.Map:
		newVal := r.randomMap(v.Type())
		r.setField(v, reflect.ValueOf(newVal))
	default:
		if isPrimitiveKind(v.Kind()) {
			pv := r.newPrimitive(v.Kind())
			rv := reflect.ValueOf(pv)
			if rv.Type().AssignableTo(v.Type()) {
				r.setField(v, rv)
			} else if rv.Type().ConvertibleTo(v.Type()) {
				r.setField(v, rv.Convert(v.Type()))
			}
		}
	}
}

// setField sets a field value, even if it's unexported, using unsafe pointer
// access.
func (*random) setField(field reflect.Value, value reflect.Value) {
	if field.CanSet() {
		field.Set(value)
	} else {
		// Use unsafe pointer to set unexported fields
		// #nosec G103 -- This is intentional for testing purposes
		reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().Set(value)
	}
}

// randomString generates a random string of random length up to r.len.
func (r *random) randomString() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	ln := r.rand.Intn(r.length-1) + 1
	b := make([]byte, ln)
	for i := range b {
		b[i] = chars[r.rand.Intn(len(chars))]
	}
	return string(b)
}
