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

type WithDeepCopy struct {
	Value int
}

func (d *WithDeepCopy) DeepCopy() any {
	if d == nil {
		return nil
	}
	out := new(WithDeepCopy)
	*out = *d
	return out
}

type WithDeepCopyObject struct {
	Value int
}

func (d *WithDeepCopyObject) DeepCopyObject() any {
	if d == nil {
		return nil
	}
	out := new(WithDeepCopyObject)
	*out = *d
	return out
}

type WithoutDeepCopy struct {
	Value int
}

type DeepCopyParams struct {
	test.DeepCopyParams
	setup mock.SetupFunc
}

var deepCopyTestCases = map[string]DeepCopyParams{
	"deep-copy-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: (*WithDeepCopy)(nil),
		},
	},
	"deep-copy": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &WithDeepCopy{},
		},
	},
	"deep-copy-object-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: (*WithDeepCopyObject)(nil),
		},
	},
	"deep-copy-object": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &WithDeepCopyObject{},
		},
	},
	"no-method-nil": {
		DeepCopyParams: test.DeepCopyParams{
			Value: nil,
		},
		setup: test.Fatalf("no deep copy method [%T]", nil),
	},
	"no-method": {
		DeepCopyParams: test.DeepCopyParams{
			Value: &WithoutDeepCopy{Value: 6},
		},
		setup: test.Fatalf("no deep copy method [%T]", &WithoutDeepCopy{Value: 6}),
	},
}

func TestDeepCopy(t *testing.T) {
	test.Map(t, deepCopyTestCases).
		Run(func(t test.Test, param DeepCopyParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			test.DeepCopy(42, 5, 20)(t, param.DeepCopyParams)
		})
}
