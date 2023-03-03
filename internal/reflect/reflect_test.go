package reflect_test

import (
	"errors"
	"testing"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/mock"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/test"
)

var (
	testchan    = make(chan test.Test)
	testint     = 1
	testfloat   = 1.0
	testcomplex = complex(1.0, 1.0)
	teststring  = "value"
	testmap     = map[string]string{"value": "value"}
	teststruct  = ExportParams{Value: "value"}
	testarray   = [1]string{"value"}
	testslice   = []string{"value"}
)

func FuncTest() {}

type FindArgOfParams struct {
	name   string
	value  any
	deflt  any
	expect any
}

type BoolParams struct {
	value bool
}

type IntParams struct {
	value int
}

type StringParams struct {
	value string
}

type ExportParams struct {
	Value string
}

type StructParams struct {
	value BoolParams
}

var testFindArgOfParams = map[string]FindArgOfParams{
	"no struct": {
		name:   "value",
		value:  "string",
		deflt:  true,
		expect: true,
	},

	"bool value": {
		name:   "any",
		value:  true,
		deflt:  false,
		expect: true,
	},

	"bool found": {
		name:   "value",
		value:  BoolParams{value: true},
		deflt:  false,
		expect: true,
	},

	"bool fallback": {
		name:   "fallback",
		value:  BoolParams{value: true},
		deflt:  true,
		expect: true,
	},

	"bool notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  true,
		expect: true,
	},

	"int value": {
		name:   "any",
		value:  2,
		deflt:  1,
		expect: 2,
	},

	"int found": {
		name:   "value",
		value:  IntParams{value: 2},
		deflt:  1,
		expect: 2,
	},

	"int fallback": {
		name:   "fallback",
		value:  IntParams{value: 2},
		deflt:  1,
		expect: 2,
	},

	"int notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  1,
		expect: 1,
	},

	"string value": {
		name:   "any",
		value:  "value",
		deflt:  "default",
		expect: "value",
	},

	"string found": {
		name:   "value",
		value:  StringParams{value: "value"},
		deflt:  "default",
		expect: "value",
	},

	"string fallback": {
		name:   "fallback",
		value:  StringParams{value: "fallback"},
		deflt:  "default",
		expect: "fallback",
	},

	"string notfound": {
		name:   "notfound",
		value:  BoolParams{},
		deflt:  "default",
		expect: "default",
	},

	"export found": {
		name:   "value",
		value:  ExportParams{Value: "value"},
		deflt:  "default",
		expect: "value",
	},

	"export fallback": {
		name:   "fallback",
		value:  ExportParams{Value: "fallback"},
		deflt:  "default",
		expect: "fallback",
	},

	"export notfound": {
		name:   "notfound",
		value:  BoolParams{},
		deflt:  "default",
		expect: "default",
	},

	"struct found": {
		name: "value",
		value: StructParams{
			value: BoolParams{value: true},
		},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: true},
	},

	"struct fallback": {
		name: "fallback",
		value: StructParams{
			value: BoolParams{value: true},
		},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: true},
	},

	"struct notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: false},
	},
}

func TestFindArgOf(t *testing.T) {
	test.Map(t, testFindArgOfParams).
		Run(func(t test.Test, param FindArgOfParams) {
			// When
			value := reflect.FindArgOf(param.value, param.deflt, param.name)

			// Then
			assert.Equal(t, param.expect, value)
		})
}

type ArgOfParams struct {
	value  reflect.Value
	expect any
}

var testArgOfParams = map[string]ArgOfParams{
	"nil": {
		value:  reflect.ValueOf(nil),
		expect: nil,
	},

	"bool": {
		value:  reflect.ValueOf(true),
		expect: true,
	},
	"int": {
		value:  reflect.ValueOf(int(1)),
		expect: int(1),
	},
	"int8": {
		value:  reflect.ValueOf(int8(1)),
		expect: int8(1),
	},
	"int16": {
		value:  reflect.ValueOf(int16(1)),
		expect: int16(1),
	},
	"int32": {
		value:  reflect.ValueOf(int32(1)),
		expect: int32(1),
	},
	"int64": {
		value:  reflect.ValueOf(int64(1)),
		expect: int64(1),
	},
	"intptr": {
		value:  reflect.ValueOf(&testint),
		expect: &testint,
	},
	"uint": {
		value:  reflect.ValueOf(uint(1)),
		expect: uint(1),
	},
	"uint8": {
		value:  reflect.ValueOf(uint8(1)),
		expect: uint8(1),
	},
	"uint16": {
		value:  reflect.ValueOf(uint16(1)),
		expect: uint16(1),
	},
	"uint32": {
		value:  reflect.ValueOf(uint32(1)),
		expect: uint32(1),
	},
	"uint64": {
		value:  reflect.ValueOf(uint64(1)),
		expect: uint64(1),
	},

	"float32": {
		value:  reflect.ValueOf(float32(1)),
		expect: float32(1),
	},
	"float64": {
		value:  reflect.ValueOf(float64(1)),
		expect: float64(1),
	},
	"floatptr": {
		value:  reflect.ValueOf(&testfloat),
		expect: &testfloat,
	},
	"complex64": {
		value:  reflect.ValueOf(complex64(testcomplex)),
		expect: complex64(testcomplex),
	},
	"complex128": {
		value:  reflect.ValueOf(testcomplex),
		expect: testcomplex,
	},
	"string": {
		value:  reflect.ValueOf("value"),
		expect: "value",
	},
	"stringptr": {
		value:  reflect.ValueOf(&teststring),
		expect: &teststring,
	},

	"array": {
		value:  reflect.ValueOf(testarray),
		expect: testarray,
	},
	"slice": {
		value:  reflect.ValueOf(testslice),
		expect: testslice,
	},
	"map": {
		value:  reflect.ValueOf(testmap),
		expect: testmap,
	},
	"struct": {
		value:  reflect.ValueOf(teststruct),
		expect: teststruct,
	},
	"chan": {
		value:  reflect.ValueOf(testchan),
		expect: testchan,
	},
	"func": {
		value:  reflect.ValueOf(FuncTest),
		expect: FuncTest,
	},
}

func TestArgOf(t *testing.T) {
	test.Map(t, testArgOfParams).
		Run(func(t test.Test, param ArgOfParams) {
			// When
			arg := reflect.ArgOf(param.value)

			// Then
			switch {
			case !param.value.IsValid():
				assert.Equal(t, param.expect, arg)
			case param.value.Type().Kind() == reflect.Func:
				assert.NotNil(t, arg)
				assert.True(t, reflect.TypeOf(arg).Kind() == reflect.Func)
			default:
				assert.Equal(t, param.expect, arg)
			}
		})
}

type ArgsOfParams struct {
	values []reflect.Value
	args   []any
}

var testArgsOfParams = map[string]ArgsOfParams{
	"values-nil": {
		values: nil,
		args:   nil,
	},
	"values-empty": {
		values: []reflect.Value{},
		args:   nil,
	},

	"nil": {
		values: []reflect.Value{
			reflect.ValueOf(nil), reflect.ValueOf(nil),
		},
		args: []any{nil, nil},
	},

	"bool": {
		values: []reflect.Value{
			reflect.ValueOf(true), reflect.ValueOf(false),
		},
		args: []any{true, false},
	},
	"int": {
		values: []reflect.Value{
			reflect.ValueOf(1), reflect.ValueOf(2),
		},
		args: []any{1, 2},
	},
	"string": {
		values: []reflect.Value{
			reflect.ValueOf("value"), reflect.ValueOf("other"),
		},
		args: []any{"value", "other"},
	},
	"stringptr": {
		values: []reflect.Value{
			reflect.ValueOf(&teststring), reflect.ValueOf(&teststring),
		},
		args: []any{&teststring, &teststring},
	},

	"array": {
		values: []reflect.Value{
			reflect.ValueOf(testarray), reflect.ValueOf(testarray),
		},
		args: []any{testarray, testarray},
	},
	"slice": {
		values: []reflect.Value{
			reflect.ValueOf(testslice), reflect.ValueOf(testslice),
		},
		args: []any{testslice, testslice},
	},
	"map": {
		values: []reflect.Value{
			reflect.ValueOf(testmap), reflect.ValueOf(testmap),
		},
		args: []any{testmap, testmap},
	},
	"struct": {
		values: []reflect.Value{
			reflect.ValueOf(teststruct), reflect.ValueOf(teststruct),
		},
		args: []any{teststruct, teststruct},
	},
	"chan": {
		values: []reflect.Value{
			reflect.ValueOf(testchan),
		},
		args: []any{testchan},
	},
}

func TestArgsOf(t *testing.T) {
	test.Map(t, testArgsOfParams).
		Run(func(t test.Test, param ArgsOfParams) {
			// When
			args := reflect.ArgsOf(param.values...)

			// Then
			assert.Equal(t, param.args, args)
		})
}

type ValuesParams struct {
	setup  mock.SetupFunc
	call   any
	args   []any
	expect test.Expect
}

var testValuesInParams = map[string]ValuesParams{
	"args-nil": {
		setup: test.Panic("not enough arguments"),
		call:  func(any) {},
		args:  nil,
	},
	"args-nil-var": {
		call:   func(...any) {},
		args:   nil,
		expect: test.Success,
	},

	"nil": {
		call:   func(any, any) {},
		args:   []any{nil, nil},
		expect: test.Success,
	},
	"nil-less": {
		setup: test.Panic("not enough arguments"),
		call:  func(any, any) {},
		args:  []any{nil},
	},
	"nil-more": {
		setup: test.Panic("too many arguments"),
		call:  func(any) {},
		args:  []any{nil, nil},
	},
	"nil-var": {
		call:   func(any, ...any) {},
		args:   []any{nil},
		expect: test.Success,
	},
	"nil-var-less": {
		setup: test.Panic("not enough arguments"),
		call:  func(any, ...any) {},
		args:  []any{},
	},
	"nil-var-first": {
		call:   func(any, ...any) {},
		args:   []any{nil, nil},
		expect: test.Success,
	},
	"nil-var-more": {
		call:   func(any, ...any) {},
		args:   []any{nil, nil, nil},
		expect: test.Success,
	},

	"match": {
		call:   func(bool, int, string, ExportParams) {},
		args:   []any{true, 1, "value", teststruct},
		expect: test.Success,
	},
	"match-var": {
		call:   func(bool, int, string, ExportParams, ...any) {},
		args:   []any{true, 1, "value", teststruct},
		expect: test.Success,
	},
	"match-var-less": {
		setup: test.Panic("not enough arguments"),
		call:  func(bool, int, string, ExportParams) {},
		args:  []any{true, 1, "value"},
	},
	"match-var-more": {
		call:   func(bool, int, string, ExportParams, ...any) {},
		args:   []any{true, 1, "value", teststruct, true, 1, "value"},
		expect: test.Success,
	},

	"miss": {
		setup: test.Panic(reflect.ErrInvalidType(
			1, reflect.TypeOf(1), reflect.TypeOf(""))),
		call: func(bool, int) {},
		args: []any{true, "value"},
	},
	"miss-var": {
		setup: test.Panic(reflect.ErrInvalidType(
			1, reflect.TypeOf(true), reflect.TypeOf(0))),
		call: func(bool, ...bool) {},
		args: []any{true, 1, "value"},
	},
}

func TestValuesIn(t *testing.T) {
	test.Map(t, testValuesInParams).
		Run(func(t test.Test, param ValuesParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)
			ftype := reflect.TypeOf(param.call)

			// When
			values := reflect.ValuesIn(ftype, param.args...)

			// Then (reflect.Values are not comparable)
			assert.Equal(t, param.args, reflect.ArgsOf(values...))
		})
}

var testValuesOutParams = map[string]ValuesParams{
	"method-nil": {
		call:   func() {},
		args:   nil,
		expect: test.Success,
	},
	"method-empty": {
		call: func() {},
		args: []any{},
	},
	"method-more": {
		setup: test.Panic("too many arguments"),
		call:  func() {},
		args:  []any{nil},
	},

	"nil": {
		call:   func() (any, any) { return nil, nil },
		args:   []any{nil, nil},
		expect: test.Success,
	},
	"nil-less": {
		setup: test.Panic("not enough arguments"),
		call:  func() (any, any) { return nil, nil },
		args:  []any{nil},
	},
	"nil-more": {
		setup: test.Panic("too many arguments"),
		call:  func() any { return nil },
		args:  []any{nil, nil},
	},

	"match": {
		call: func() (bool, int, string, ExportParams) {
			return true, 1, "value", teststruct
		},
		args:   []any{true, 1, "value", teststruct},
		expect: test.Success,
	},

	"miss": {
		setup: test.Panic(reflect.ErrInvalidType(
			1, reflect.TypeOf(true), reflect.TypeOf(1))),
		call: func() (bool, bool) { return true, true },
		args: []any{true, 1},
	},
}

func TestValuesOut(t *testing.T) {
	test.Map(t, testValuesOutParams).
		Run(func(t test.Test, param ValuesParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)
			ftype := reflect.TypeOf(param.call)

			// When
			values := reflect.ValuesOut(ftype, param.args...)

			// Then (reflect.Values are not comparable)
			assert.Equal(t, param.args, reflect.ArgsOf(values...))
		})
}

type AnyFuncOfParams struct {
	args     int
	variadic bool
	expect   any
}

var testAnyFuncOfParams = map[string]AnyFuncOfParams{
	"args-0": {
		args: 0, variadic: false, expect: func() {},
	},
	"args-1": {
		args: 1, variadic: false, expect: func(any) {},
	},
	"args-1-var": {
		args: 1, variadic: true, expect: func(...any) {},
	},
	"args-2": {
		args: 2, variadic: false, expect: func(any, any) {},
	},
	"args-2-var": {
		args: 2, variadic: true, expect: func(any, ...any) {},
	},
	"args-4": {
		args: 4, variadic: false, expect: func(any, any, any, any) {},
	},
	"args-4-var": {
		args: 4, variadic: true, expect: func(any, any, any, ...any) {},
	},
}

func TestAnyFuncOf(t *testing.T) {
	test.Map(t, testAnyFuncOfParams).
		Run(func(t test.Test, param AnyFuncOfParams) {
			// Given

			// When
			ftype := reflect.AnyFuncOf(param.args, param.variadic)

			// Then
			assert.Equal(t, reflect.TypeOf(param.expect), ftype)
		})
}

type BaseFuncOfParams struct {
	call    any
	in, out int
	expect  any
}

var testBaseFuncOfParams = map[string]BaseFuncOfParams{
	"in-0-out-0": {
		call:   func() {},
		expect: func() {},
	},
	"in-1-out-0": {
		call:   func(any) {},
		expect: func(any) {},
	},
	"in-2-out-0": {
		call:   func(any, string) {},
		expect: func(any, string) {},
	},

	"in-2(1)-out-0": {
		in:     1,
		call:   func(any, string) {},
		expect: func(string) {},
	},
	"in-2(0)-out-0": {
		in:     2,
		call:   func(any, string) {},
		expect: func() {},
	},

	"in-1-out-1": {
		call:   func(any) int { return 1 },
		expect: func(any) int { return 1 },
	},
	"in-1-out-2": {
		call:   func(any) (int, error) { return 1, nil },
		expect: func(any) (int, error) { return 1, nil },
	},
	"in-2-out-1": {
		call:   func(any, string) int { return 1 },
		expect: func(any, string) int { return 1 },
	},
	"in-2-out-2": {
		call:   func(any, string) (int, error) { return 1, nil },
		expect: func(any, string) (int, error) { return 1, nil },
	},

	"in-1(0)-out-1(0)": {
		in: 1, out: 1,
		call:   func(any) int { return 1 },
		expect: func() {},
	},
	"in-1(0)-out-2(1)": {
		in: 1, out: 1,
		call:   func(any) (int, error) { return 1, nil },
		expect: func() error { return nil },
	},
	"in-2(0)-out-1(0)": {
		in: 2, out: 2,
		call:   func(any, string) int { return 1 },
		expect: func() {},
	},
	"in-2(0)-out-2(0)": {
		in: 2, out: 2,
		call:   func(any, string) (int, error) { return 1, nil },
		expect: func() {},
	},
}

func TestBaseFuncOf(t *testing.T) {
	test.Map(t, testBaseFuncOfParams).
		Run(func(t test.Test, param BaseFuncOfParams) {
			// Given
			ctype := reflect.TypeOf(param.call)

			// When
			ftype := reflect.BaseFuncOf(ctype, param.in, param.out)

			// Then
			assert.Equal(t, reflect.TypeOf(param.expect), ftype)
		})
}

type MakeFuncOfParams struct {
	call       any
	args       []any
	result     []any
	expectArgs []any
}

var testMakeFuncOfParams = map[string]MakeFuncOfParams{
	"in-0-out-0": {
		call: func() {},
	},
	"in-1-out-1": {
		call:       func(any) int { return 1 },
		args:       []any{teststruct},
		result:     []any{1},
		expectArgs: []any{teststruct},
	},
	"in-2-out-2": {
		call:       func(any, string) (int, error) { return 1, nil },
		args:       []any{teststruct, teststring},
		result:     []any{1, errors.New("any")},
		expectArgs: []any{teststruct, teststring},
	},
	"in-2-var-out-2": {
		call:       func(any, ...string) (int, error) { return 1, nil },
		args:       []any{teststruct, teststring, "string-2", "string-3"},
		result:     []any{1, errors.New("any")},
		expectArgs: []any{teststruct, []string{teststring, "string-2", "string-3"}},
	},
}

func TestMakeFuncOf(t *testing.T) {
	test.Map(t, testMakeFuncOfParams).
		Run(func(t test.Test, param MakeFuncOfParams) {
			// Given
			ctype := reflect.TypeOf(param.call)

			// When
			call := reflect.MakeFuncOf(ctype,
				func(args []reflect.Value) []reflect.Value {
					assert.Equal(t, param.expectArgs, reflect.ArgsOf(args...))
					return reflect.ValuesOut(ctype, param.result...)
				})

			// Then
			ftype := reflect.TypeOf(call)
			assert.Equal(t, ctype, ftype)

			// When
			result := reflect.ArgsOf(reflect.ValueOf(call).Call(
				reflect.ValuesIn(ftype, param.args...),
			)...)

			// Then
			assert.Equal(t, param.result, result)
		})
}
