package test_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

//revive:disable:max-public-structs // relaxed in testing.

// Named types for testing.
type (
	TestSlice  []string
	TestMap    map[string]int
	TestStruct struct {
		name string
		id   int
	}
	TestFunc func(*any) any
)

var testFunc TestFunc = func(a *any) any { return a }

type MustParams struct {
	setup  mock.SetupFunc
	arg    any
	err    error
	expect any
}

var mustTestCases = map[string]MustParams{
	"nil": {},
	"string": {
		arg:    "value",
		expect: "value",
	},
	"integer": {
		arg:    1,
		expect: 1,
	},
	"float": {
		arg:    3.14,
		expect: 3.14,
	},
	"bool true": {
		arg:    true,
		expect: true,
	},
	"bool false": {
		arg:    false,
		expect: false,
	},
	"slice": {
		arg:    []string{"a", "b", "c"},
		expect: []string{"a", "b", "c"},
	},
	"map": {
		arg:    map[string]int{"key": 42},
		expect: map[string]int{"key": 42},
	},
	"struct": {
		arg:    TestStruct{name: "test", id: 123},
		expect: TestStruct{name: "test", id: 123},
	},
	"pointer": {
		arg:    &TestStruct{name: "pointer", id: 456},
		expect: &TestStruct{name: "pointer", id: 456},
	},
	"function": {
		arg:    testFunc,
		expect: testFunc,
	},
	"named slice": {
		arg:    TestSlice{"x", "y", "z"},
		expect: TestSlice{"x", "y", "z"},
	},
	"named map": {
		arg:    TestMap{"foo": 1, "bar": 2},
		expect: TestMap{"foo": 1, "bar": 2},
	},
	"zero value int": {
		arg:    0,
		expect: 0,
	},
	"zero value string": {
		arg:    "",
		expect: "",
	},
	"empty slice": {
		arg:    []string{},
		expect: []string{},
	},
	"empty map": {
		arg:    map[string]int{},
		expect: map[string]int{},
	},
	"error": {
		setup:  test.Panic(assert.AnError),
		err:    assert.AnError,
		expect: nil,
	},
	"nil error with value": {
		arg:    "success",
		err:    nil,
		expect: "success",
	},
}

func TestMust(t *testing.T) {
	test.Map(t, mustTestCases).Run(func(t test.Test, param MustParams) {
		// Given
		mock.NewMocks(t).Expect(param.setup)

		// When
		result := test.Must(param.arg, param.err)

		// Then
		if strings.Contains(t.Name(), "function") {
			// For functions, just verify the result is not nil (functions can't be compared)
			assert.NotNil(t, result)
		} else {
			assert.Equal(t, param.expect, result)
		}
	})
}

type CastParams struct {
	setup  mock.SetupFunc
	arg    any
	cast   func(arg any) any
	expect any
}

var castTestCases = map[string]CastParams{
	"int to int": {
		arg:    42,
		cast:   func(arg any) any { return test.Cast[int](arg) },
		expect: 42,
	},
	"string to string": {
		arg:    "value",
		cast:   func(arg any) any { return test.Cast[string](arg) },
		expect: "value",
	},
	"slice to slice string": {
		arg:    []string{"a", "b", "c"},
		cast:   func(arg any) any { return test.Cast[[]string](arg) },
		expect: []string{"a", "b", "c"},
	},
	"slice to slice names": {
		arg:    TestSlice{"x", "y", "z"},
		cast:   func(arg any) any { return test.Cast[TestSlice](arg) },
		expect: TestSlice{"x", "y", "z"},
	},
	"map to map": {
		arg:    TestMap{"foo": 1, "bar": 2},
		cast:   func(arg any) any { return test.Cast[TestMap](arg) },
		expect: TestMap{"foo": 1, "bar": 2},
	},
	"struct to struct": {
		arg:    TestStruct{name: "example", id: 123},
		cast:   func(arg any) any { return test.Cast[TestStruct](arg) },
		expect: TestStruct{name: "example", id: 123},
	},
	"pointer to pointer": {
		arg:    &TestStruct{name: "example", id: 123},
		cast:   func(arg any) any { return test.Cast[*TestStruct](arg) },
		expect: &TestStruct{name: "example", id: 123},
	},
	"function to function": {
		arg:    testFunc,
		cast:   func(arg any) any { return test.Cast[TestFunc](arg) },
		expect: testFunc,
	},

	// Casts to any.
	"int to any": {
		arg:    123,
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: 123,
	},
	"bool to any": {
		arg:    true,
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: true,
	},
	"string to any": {
		arg:    "hello",
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: "hello",
	},
	"slice to any int": {
		arg:    []int{1, 2, 3},
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: []int{1, 2, 3},
	},
	"slice to any named": {
		arg:    TestSlice{"a", "b", "c"},
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: TestSlice{"a", "b", "c"},
	},
	"map to any": {
		arg:    TestMap{"key": 42, "other": 100},
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: TestMap{"key": 42, "other": 100},
	},
	"struct to any": {
		arg:    TestStruct{name: "test", id: 42},
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: TestStruct{name: "test", id: 42},
	},
	"pointer to any": {
		arg:    &TestStruct{name: "pointer", id: 99},
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: &TestStruct{name: "pointer", id: 99},
	},
	"function to any": {
		arg:    fmt.Sprintf,
		cast:   func(arg any) any { return test.Cast[any](arg) },
		expect: nil, // Special case - just test that cast succeeds
	},

	// Panic cases
	"nil to any": {
		setup: test.Panic(fmt.Sprintf("cast failed [<nil>]: %v", nil)),
		arg:   nil,
		cast:  func(arg any) any { return test.Cast[any](arg) },
	},
	"string to int": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(0), "value")),
		arg:  "value",
		cast: func(arg any) any { return test.Cast[int](arg) },
	},
	"int to string": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(""), 42)),
		arg:  42,
		cast: func(arg any) any { return test.Cast[string](arg) },
	},
	"nil to string": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(""), nil)),
		arg:  nil,
		cast: func(arg any) any { return test.Cast[string](arg) },
	},
	"float to int": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(0), 3.14)),
		arg:  3.14,
		cast: func(arg any) any { return test.Cast[int](arg) },
	},
	"string to struct": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(TestStruct{}), "invalid")),
		arg:  "invalid",
		cast: func(arg any) any { return test.Cast[TestStruct](arg) },
	},
	"int to map": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(TestMap{}), 42)),
		arg:  42,
		cast: func(arg any) any { return test.Cast[TestMap](arg) },
	},
	"slice to named slice": {
		setup: test.Panic(fmt.Sprintf("cast failed [%v]: %v",
			reflect.TypeOf(TestSlice{}), []string{"too", "many"})),
		arg:  []string{"too", "many"},
		cast: func(arg any) any { return test.Cast[TestSlice](arg) },
	},
}

func TestCast(t *testing.T) {
	test.Map(t, castTestCases).Run(func(t test.Test, param CastParams) {
		// Given
		mock.NewMocks(t).Expect(param.setup)

		// When
		result := param.cast(param.arg)

		// Then
		if strings.Contains(t.Name(), "function") {
			assert.NotNil(t, result)
		} else if strings.Contains(t.Name(), "pointer to any") {
			expect := param.expect.(*TestStruct)
			result := result.(*TestStruct)
			assert.Equal(t, expect.name, result.name)
			assert.Equal(t, expect.id, result.id)
		} else if param.expect == nil {
			assert.NotNil(t, result)
		} else {
			assert.Equal(t, param.expect, result)
		}
	})
}

type PtrParams struct {
	value any
}

var ptrTestCases = map[string]PtrParams{
	// Primitive types
	"bool true":  {value: true},
	"bool false": {value: false},

	// Integer types
	"int":          {value: 42},
	"int zero":     {value: 0},
	"int negative": {value: -123},
	"int8":         {value: int8(127)},
	"int16":        {value: int16(32767)},
	"int32":        {value: int32(2147483647)},
	"int64":        {value: int64(9223372036854775807)},

	// Unsigned integer types
	"uint":   {value: uint(42)},
	"uint8":  {value: uint8(255)},
	"uint16": {value: uint16(65535)},
	"uint32": {value: uint32(4294967295)},
	"uint64": {value: uint64(18446744073709551615)},
	"byte":   {value: byte(255)},
	"rune":   {value: rune('A')},

	// Floating point types
	"float32":          {value: float32(3.14)},
	"float32 zero":     {value: float32(0.0)},
	"float64":          {value: 3.141592653589793},
	"float64 negative": {value: -2.718281828},

	// Complex types
	"complex64":  {value: complex64(1 + 2i)},
	"complex128": {value: complex(3.0, 4.0)},

	// String types
	"string":         {value: "hello world"},
	"string empty":   {value: ""},
	"string unicode": {value: "Hello, 🌍"},

	// Slice literals
	"slice int":    {value: []int{1, 2, 3}},
	"slice string": {value: []string{"a", "b", "c"}},
	"slice empty":  {value: []int{}},
	"slice nil":    {value: []int(nil)},

	// Map literals
	"map string int": {value: map[string]int{"one": 1, "two": 2}},
	"map empty":      {value: map[string]int{}},
	"map nil":        {value: map[string]int(nil)},

	// Struct literals
	"struct":           {value: TestStruct{name: "test", id: 42}},
	"struct zero":      {value: TestStruct{}},
	"struct anonymous": {value: struct{ X int }{X: 10}},

	// Named types
	"named slice": {value: TestSlice{"x", "y", "z"}},
	"named map":   {value: TestMap{"foo": 1, "bar": 2}},

	// Pointer types
	"pointer to int":    {value: test.Ptr(42)},
	"pointer to string": {value: test.Ptr("value")},
	"pointer to struct": {value: &TestStruct{name: "ptr", id: 99}},

	// Array literals
	"array int":    {value: [3]int{1, 2, 3}},
	"array string": {value: [2]string{"hello", "world"}},

	// Interface types
	"interface any": {value: any("interface value")},
}

func TestPtr(t *testing.T) {
	test.Map(t, ptrTestCases).Run(func(t test.Test, param PtrParams) {
		// When
		result := test.Ptr(param.value)

		// Then
		assert.Equal(t, &param.value, result)

		rvalue := reflect.ValueOf(result)
		assert.Equal(t, reflect.Ptr, rvalue.Kind())

		value := rvalue.Elem().Interface()
		assert.Equal(t, param.value, value)
	})
}

type RecoverParams struct {
	setup  any
	expect test.Expect
	panic  any
}

var recoverTestCases = map[string]RecoverParams{
	// Failure to panic.
	"no panic with nil": {},
	"no panic with string": {
		setup: "no panic",
	},
	"no panic with error": {
		setup: assert.AnError,
	},
	"no panic with pointer error": {
		setup: &assert.AnError,
	},

	// Successful recovery.
	"panic with string": {
		panic:  "a panic occurred",
		setup:  "a panic occurred",
		expect: test.Success,
	},
	"panic with error": {
		panic:  assert.AnError,
		setup:  assert.AnError,
		expect: test.Success,
	},
	"panic with same error pointer": {
		panic:  &assert.AnError,
		setup:  &assert.AnError,
		expect: test.Success,
	},
}

func TestRecover(t *testing.T) {
	test.Map(t, recoverTestCases).
		Run(func(t test.Test, param RecoverParams) {
			// Given
			defer test.Recover(t, param.setup)

			// When
			if param.panic != nil {
				panic(param.panic)
			}
		})
}

// ctx returns the current time formatted as RFC3339Nano truncated to 26
// characters to avoid excessive precision in test output.
func ctx() string {
	return fmt.Sprintf("%s [%s]", time.Now().Format(time.RFC3339Nano[0:26]), os.Args[0])
}

// main is a test main function to demonstrate the usage of `test.Main`.
func main() {
	// Check that environment variables are set correctly.
	fmt.Fprintf(os.Stderr, "%s var=%s\n", ctx(), os.Getenv("var"))
	if os.Getenv("panic") == "true" {
		fmt.Fprintf(os.Stderr, "%s var=%s\n", ctx(), os.Getenv("var"))
		panic("supposed to panic")
	}

	// Simulate some work.
	fmt.Fprintf(os.Stderr, "%s args=%v\n", ctx(), os.Args)
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "%s sleep=%s\n", ctx(), os.Args[1])
		dur, err := time.ParseDuration(os.Args[1])
		if err == nil {
			time.Sleep(dur)
		}
	}

	// Exit with given code.
	fmt.Fprintf(os.Stderr, "%s exit=%s\n", ctx(), os.Getenv("exit"))
	os.Exit(test.First(strconv.Atoi(os.Getenv("exit"))))
}

var mainTestCases = map[string]test.MainParams{
	"panic": {
		Env:      []string{"panic=true"},
		Args:     []string{"panic"},
		ExitCode: 1,
	},
	"exit-1": {
		Env:      []string{"exit=1", "panic=false"},
		Args:     []string{"exit-1"},
		ExitCode: 1,
	},
	"exit-0": {
		Env:      []string{"exit=0", "panic=true", "panic=false"},
		Args:     []string{"exit-0"},
		ExitCode: 0,
	},
	"sleep": {
		Args:     []string{"sleep", "100ms"},
		ExitCode: 0,
	},
	"deadline": {
		Args: []string{"deadline", "1s"},
		Ctx: test.First(context.WithTimeout(context.Background(),
			time.Millisecond)),
		Error:    context.DeadlineExceeded,
		ExitCode: -1,
	},
	"interrupt": {
		Args: []string{"interrupt", "1s"},
		Ctx: test.First(context.WithTimeout(context.Background(),
			500*time.Millisecond)),
		ExitCode: -1,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, mainTestCases).Run(test.Main(main))
}

func TestMainUnexpected(t *testing.T) {
	t.Setenv(test.GoTestingRunVar, "other")
	test.Param(t, test.MainParams{}).RunSeq(test.Main(main))
}

type (
	noCopySlice   []int
	deepCopySlice []int
)

func (d deepCopySlice) DeepCopy() deepCopySlice {
	if d == nil {
		return nil
	}

	copied := make(deepCopySlice, len(d))
	copy(copied, d)

	return copied
}

type (
	noCopyMap   map[string]int
	deepCopyMap map[string]int
)

func (d deepCopyMap) DeepCopy() deepCopyMap {
	if d == nil {
		return nil
	}

	copied := make(deepCopyMap)
	for k, v := range d {
		copied[k] = v
	}

	return copied
}

type noCopyStruct struct {
	Value int
}

type deepCopyStruct struct {
	Value int
}

func (d *deepCopyStruct) DeepCopy() *deepCopyStruct {
	if d == nil {
		return nil
	}

	return &deepCopyStruct{Value: d.Value}
}

type deepCopyObject struct {
	Value int
}

func (d *deepCopyObject) DeepCopyObject() *deepCopyObject {
	if d == nil {
		return nil
	}

	return &deepCopyObject{Value: d.Value}
}

// DeepCopyCasesParams defines parameters for testing DeepCopyTestCases.
type DeepCopyCasesParams struct {
	args   []any
	expect map[string]test.DeepCopyParams
}

var deepCopyTestCasesTestCases = map[string]DeepCopyCasesParams{
	"struct types": {
		args: []any{
			&noCopySlice{},
			&deepCopySlice{},
			&noCopyMap{},
			&deepCopyMap{},
			&noCopyStruct{},
			&deepCopyStruct{},
			&deepCopyObject{},
		},
		expect: map[string]test.DeepCopyParams{
			"no-copy-slice-nil": {
				Value: (*noCopySlice)(nil),
			},
			"no-copy-slice-value": {
				Value: &noCopySlice{8, 9, 1},
			},
			"deep-copy-slice-nil": {
				Value: (*deepCopySlice)(nil),
			},
			"deep-copy-slice-value": {
				Value: &deepCopySlice{6, 8},
			},
			"no-copy-map-nil": {
				Value: (*noCopyMap)(nil),
			},
			"no-copy-map-value": {
				Value: &noCopyMap{
					"6jif9":    5,
					"g21qrot5": 10,
					"hbl":      6,
				},
			},
			"deep-copy-map-nil": {
				Value: (*deepCopyMap)(nil),
			},
			"deep-copy-map-value": {
				Value: &deepCopyMap{"5srwst": 3},
			},
			"no-copy-struct-nil": {
				Value: (*noCopyStruct)(nil),
			},
			"no-copy-struct-value": {
				Value: &noCopyStruct{Value: 8},
			},
			"deep-copy-struct-nil": {
				Value: (*deepCopyStruct)(nil),
			},
			"deep-copy-struct-value": {
				Value: &deepCopyStruct{Value: 6},
			},
			"deep-copy-object-nil": {
				Value: (*deepCopyObject)(nil),
			},
			"deep-copy-object-value": {
				Value: &deepCopyObject{Value: 8},
			},
		},
	},
	"anonymous struct type": {
		args: []any{&struct{ Value int }{}},
		expect: map[string]test.DeepCopyParams{
			"struct { value int } nil": {
				Value: (*struct{ Value int })(nil),
			},
			"struct { value int } value": {
				Value: &struct{ Value int }{Value: 6},
			},
		},
	},
}

func TestDeepCopyTestCases(t *testing.T) {
	test.Map(t, deepCopyTestCasesTestCases).
		Run(func(t test.Test, param DeepCopyCasesParams) {
			// When
			cases := test.DeepCopyTestCases(42, 3, 10, param.args...)

			// Then
			assert.Equal(t, param.expect, cases)
		})
}

type DeepCopyParams struct {
	test.DeepCopyParams
	setup mock.SetupFunc
}

var deepCopyTestCases = map[string]DeepCopyParams{
	"deep-copy-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: (*deepCopyStruct)(nil),
		},
	},
	"deep-copy-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &deepCopyStruct{Value: 2},
		},
	},
	"deep-copy-object-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: (*deepCopyObject)(nil),
		},
	},
	"deep-copy-object-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &deepCopyObject{Value: 4},
		},
	},
	"deep-copy-slice-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: deepCopySlice(nil),
		},
	},
	"deep-copy-slice-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: deepCopySlice{1, 2, 3},
		},
	},
	"deep-copy-map-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: deepCopyMap(nil),
		},
	},
	"deep-copy-map-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: deepCopyMap{"key": 42},
		},
	},
	"no-copy-method-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: nil,
		},
		setup: test.Fatalf("no deep copy method [%T]", nil),
	},
	"no-copy-method-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &noCopyStruct{Value: 6},
		},
		setup: test.Fatalf("no deep copy method [%T]",
			&noCopyStruct{Value: 6}),
	},
	"anonymous-struct-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: (*struct{ Value int })(nil),
		},
		setup: test.Fatalf("no deep copy method [%T]",
			(*struct{ Value int })(nil)),
	},
	"anonymous-struct-value": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &struct{ Value int }{Value: 8},
		},
		setup: test.Fatalf("no deep copy method [%T]",
			&struct{ Value int }{Value: 8}),
	},
}

func TestDeepCopy(t *testing.T) {
	test.Map(t, deepCopyTestCases).
		Run(func(t test.Test, param DeepCopyParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			test.DeepCopy(t, param.DeepCopyParams)
		})
}
