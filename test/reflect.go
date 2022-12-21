package test

import (
	"reflect"
	"unsafe"
)

func extract[P any](param P, deflt any, names ...string) any {
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
					return getReflect(v, i)
				}
			}
			if !found {
				deflt = getReflect(v, i)
				found = true
			}
		}
	}
	return deflt
}

func getReflect(v reflect.Value, i int) any {
	vf := v.Field(i)
	if vf.CanInterface() {
		return getValue(vf)
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

func getValue(v reflect.Value) any {
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
