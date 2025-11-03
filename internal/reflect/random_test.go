package reflect_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	. "github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/test"
)

// Custom type aliases to test type conversion.
type (
	MyInt    int
	MyString string
	MyFloat  float64
)

//lint:ignore U1000 // needed by reflection.
type Simple struct {
	a bool
	b *bool
	c int
	d *int
	e float64
	f *float64
	g string
	h *string
	// Custom type aliases.
	i MyInt
	j *MyInt
	k MyString
	l *MyString
	m MyFloat
	n *MyFloat
}

//lint:ignore U1000 // needed by reflection.
type Complex struct {
	st  Simple
	stp *Simple
	sl  []Simple
	slp []*Simple
	m   map[string]Simple
	mp  map[string]*Simple
}

// check verifies that a value is "random" by checking it's not the zero
// value. Exception: bool values are allowed to be either true or false
// For composite types (slices, maps, structs), it recursively validates all
// elements/fields.
func check(t test.Test, value any) {
	assert.NotNil(t, value)
	if value == nil {
		return
	}

	v := reflect.ValueOf(value)

	// Bool is always allowed to be zero or non-zero
	if v.Kind() == reflect.Bool {
		return
	}

	// For composite types, recursively check elements/fields
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		assert.False(t, v.IsZero())
		for i := range v.Len() {
			check(t, v.Index(i).Interface())
		}
	case reflect.Map:
		assert.False(t, v.IsZero())
		for _, key := range v.MapKeys() {
			check(t, key.Interface())
			check(t, v.MapIndex(key).Interface())
		}
	case reflect.Struct:
		// For structs, check all fields (both exported and unexported)
		checkStruct(v, t)
	case reflect.Ptr:
		assert.False(t, v.IsZero())
		check(t, v.Elem().Interface())
	default:
		// For all other primitive types, check they're not zero
		assert.False(t, v.IsZero())
	}
}

// checkStruct recursively checks all fields of a struct value.
func checkStruct(v reflect.Value, t test.Test) {
	for i := range v.NumField() {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		// Skip unexported bool fields as they may be zero
		if field.Kind() == reflect.Bool && !fieldType.IsExported() {
			continue
		}

		// For unexported fields, use unsafe access to get the value
		var value any
		if fieldType.IsExported() {
			value = field.Interface()
		} else {
			// Make value addressable if needed
			if !field.CanAddr() {
				// Create an addressable copy
				ptr := reflect.New(v.Type())
				ptr.Elem().Set(v)
				field = ptr.Elem().Field(i)
			}
			// Use unsafe pointer to access unexported field value
			value = reflect.NewAt(field.Type(),
				unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
		}
		check(t, value)
	}
}

type RandomParams struct {
	value  any
	expect test.Expect
	check  func(t test.Test, value any)
}

var randomTestCases = map[string]RandomParams{
	"nil value": {
		value:  nil,
		expect: test.Success,
		check: func(t test.Test, value any) {
			assert.Nil(t, value)
		},
	},
	"nil pointer": {
		value:  (*Simple)(nil),
		expect: test.Success,
		check:  check,
	},
	"bool": {
		value:  false,
		expect: test.Success,
		check:  check,
	},
	"int": {
		value:  int(0),
		expect: test.Success,
		check:  check,
	},
	"int8": {
		value:  int8(0),
		expect: test.Success,
		check:  check,
	},
	"int16": {
		value:  int16(0),
		expect: test.Success,
		check:  check,
	},
	"int32": {
		value:  int32(0),
		expect: test.Success,
		check:  check,
	},
	"int64": {
		value:  int64(0),
		expect: test.Success,
		check:  check,
	},
	"uint": {
		value:  uint(0),
		expect: test.Success,
		check:  check,
	},
	"uint8": {
		value:  uint8(0),
		expect: test.Success,
		check:  check,
	},
	"uint16": {
		value:  uint16(0),
		expect: test.Success,
		check:  check,
	},
	"uint32": {
		value:  uint32(0),
		expect: test.Success,
		check:  check,
	},
	"uint64": {
		value:  uint64(0),
		expect: test.Success,
		check:  check,
	},
	"uintptr": {
		value:  uintptr(0),
		expect: test.Success,
		check:  check,
	},
	"float32": {
		value:  float32(0),
		expect: test.Success,
		check:  check,
	},
	"float64": {
		value:  float64(0),
		expect: test.Success,
		check:  check,
	},
	"complex64": {
		value:  complex64(0),
		expect: test.Success,
		check:  check,
	},
	"complex128": {
		value:  complex128(0),
		expect: test.Success,
		check:  check,
	},
	"string": {
		value:  "",
		expect: test.Success,
		check:  check,
	},
	"slice int": {
		value:  []int{},
		expect: test.Success,
		check:  check,
	},
	"slice string": {
		value:  []string{},
		expect: test.Success,
		check:  check,
	},
	"map string-int": {
		value:  map[string]int{},
		expect: test.Success,
		check:  check,
	},
	"struct value": {
		value:  Simple{},
		expect: test.Success,
		check:  check,
	},
	"struct simple": {
		value:  &Simple{},
		expect: test.Success,
		check:  check,
	},
	"struct complex": {
		value: &struct {
			Ints  []int
			Names []string
			KV    map[string]int
		}{},
		expect: test.Success,
		check:  check,
	},
	"struct nested": {
		value:  &Complex{},
		expect: test.Success,
		check:  check,
	},
	"type channel": {
		value:  make(chan int),
		expect: test.Success,
		check:  check,
	},
	"type func": {
		value:  func() {},
		expect: test.Success,
		check:  check,
	},
}

func TestRandom(t *testing.T) {
	test.Map(t, randomTestCases).
		Run(func(t test.Test, p RandomParams) {
			// Given
			rand := NewRandom(42, 5, 20)

			// When
			result := rand.Random(p.value)

			// Then
			p.check(t, result)
		})
}
