package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkrop/go-testing/internal/reflect"
)

// Must is a convenience method returning the value of the first argument and
// that panics on any error in the second argument using the provided error.
// The method allows to write concise test setup code.
func Must[T any](arg T, err error) T {
	if err != nil {
		panic(err)
	}
	return arg
}

// Cast is a convenience function to cast the given argument to the specified
// type or panic if the cast fails. The method allows to write concise test
// setup code granting meaningful type checks.
func Cast[T any](arg any) T {
	val, ok := arg.(T)
	if !ok {
		panic(fmt.Sprintf("cast failed [%T]: %v", val, arg))
	}
	return val
}

// Ptr is a convenience function to obtain the pointer to the given value.
// This is particularly useful to create pointers to literal values in test
// setup code, e.g., `test.Ptr(42)` or `test.Ptr("value")`.
func Ptr[T any](v T) *T {
	return &v
}

// First is a convenience function to return the first argument and ignore all
// others arguments. The method allows to write concise test setup code.
func First[T any](arg T, _ ...any) T { return arg }

// Recover is a convenience method to be used in deferred calls to verify that
// a panic occurred with the expected panic response. The function creates a
// test failure if no panic occurred or the panic response does not match the
// expected value.
func Recover(t Test, expect any) {
	// revive:disable-next-line:defer // caller is expected to use defer.
	if actual := recover(); actual != nil {
		assert.Equal(t, expect, actual)
	} else {
		assert.Fail(t, "did not panic: %#v", expect)
	}
}

// TODO: consider following convenience methods:
//
// // Check is a convenience method that returns the second argument and swallows
// // the first used to focus a test on the second.
// func Check[T any](swallowed any, check T) T {
// 	return check
// }

// // Ok is a convenience method to check whether the second boolean argument is
// // `true` while returning the first argument. If the boolean argument is
// // `false`, the method panics.
// func Ok[T any](result T, ok bool) T {
// 	if !ok {
// 		panic("bool not okay")
// 	}
// 	return result
// }

// MainParams provides the test parameters for testing a `main`-method.
type MainParams struct {
	// Ctx is the context to control the test process, e.g. for deadlines and
	// cancellations. If not set, a background context is used with unlimited
	// duration.
	Ctx context.Context //nolint:containedctx // only in test parameters.

	// Args are the command line arguments to provide to the test process. The
	// first argument is usually the program name, typically `os.Args[0]`. If
	// not provided, the `main`-method is called with no arguments at all.
	Args []string

	// Env are the additional environment variables that are provided to the
	// spawned test process. The existing environment variables are inherited
	// and an additional `GO_TESTING_TEST` variable is set to select the test
	// case.
	Env []string

	// ExitCode is the expected exit code when running the test process. If
	// the exit code is not `0`, the error must be of type `*exec.ExitError`.
	ExitCode int

	// Error is the expected error when running the test process. This is only
	// used in edge cases when the test error is not of type `*exec.ExitError`.
	Error error
}

// GoTestingRunVar is the environment variable used to signal the new process
// to execute the `main` method instead of spawning a new test process.
const GoTestingRunVar = "GO_TESTING_RUN"

// Main creates a test function that runs the given `main`-method in a separate
// test process to protect the test execution from `os.Exit` calls while allowing
// to capture and check the exit code against the expectation. The following
// example demonstrates how to use this method to test a `main`-method:
//
//	 mainTestCases := map[string]test.MainParam{
//		 "with args": {
//			 Args: []string{"mock", "arg1", "arg2"},
//			 Env:  []string{"VAR=value"},
//			 ExitCode: 0,
//		 },
//	 }
//
//	 func Main(t *testing.T) {
//		 test.Map(t, mainTestCases).Run(test.Main(main))
//	 }
//
// If the test process is expected to run longer than the default test timeout,
// a context with timeout can be provided to interrupt the test process in time.
// This e.g. can be done as follows using `test.First` to ignore the cancelFunc:
//
//	Ctx: test.First(context.WithTimeout(context.Background(), time.Second))
func Main(main func()) func(t Test, param MainParams) {
	return func(t Test, param MainParams) {
		// Switch to execute main function in test process.
		if name := os.Getenv(GoTestingRunVar); name != "" {
			// Ensure only expected test is running.
			if name == t.Name() {
				os.Args = param.Args
				main()
				require.Fail(t, "os-exit not called")
			}
			// Skip unexpected tests.
			return
		}

		// Prepare environment for the test process.
		ctx := context.Background()
		if param.Ctx != nil {
			ctx = param.Ctx
		}

		// #nosec G204 -- secured by calling only the test instance.
		cmd := exec.CommandContext(ctx, os.Args[0],
			"-test.run="+t.(*Context).t.Name())

		// No stdout to allow propagation of coverage results.
		cmd.Stdin, cmd.Stderr = os.Stdin, os.Stderr
		cmd.Env = append(os.Environ(), append(param.Env,
			GoTestingRunVar+"="+t.Name())...)

		if err := cmd.Run(); err != nil || param.ExitCode != 0 {
			errExit := &exec.ExitError{}
			if errors.As(err, &errExit) {
				require.Equal(t, param.ExitCode, errExit.ExitCode())
			} else if err != nil {
				require.Equal(t, param.Error, err)
			}
		}
		require.Equal(t, param.ExitCode, cmd.ProcessState.ExitCode())
	}
}

// DeepCopyParams provides test parameters for testing `DeepCopy*` functions
// generated by `k8s.io/code-generator/cmd/deepcopy-gen`, that unfortunately
// are part of the type system and thus should be unit tested for coverage.
type DeepCopyParams struct {
	// Value is a template value used to generate random non-zero test values
	// for testing the `DeepCopy*` functions.
	Value any
}

// typeToTestName converts a reflect.Type into a human readable test case name.
// The name is derived from the base (non-pointer) type's CamelCase identifier
// converted to a hyphen-separated lower-case string. For unnamed types the
// string representation of the type is used as a fallback.
func typeToTestName(typ reflect.Type) string {
	raw := typ.Name()
	if raw == "" {
		raw = typ.String()
	}
	runes := []rune(raw)

	last := byte('-')
	var b strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(runes[i-1]) ||
			(i+1 < len(runes) && unicode.IsLower(runes[i+1]))) &&
			(unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
		if r == ' ' {
			last = byte(' ')
		}
	}
	b.WriteByte(last)
	return b.String()
}

// DeepCopyTestCases creates the test cases from a list of types to be tested
// either given as value or as nil pointer. For each type two test cases are
// created:
//
// * A non-nil random value, and
// * A nil pointer of the type.
//
// The types must implement either the `DeepCopy` or the `DeepCopyObject`
// method generated by `k8s.io/code-generator/cmd/deepcopy-gen` to succeed.
func DeepCopyTestCases(
	seed int64, size, length int, args ...any,
) map[string]DeepCopyParams {
	random := reflect.NewRandom(seed, size, length)
	cases := make(map[string]DeepCopyParams)
	for _, arg := range args {
		base := reflect.TypeOf(arg)
		for base.Kind() == reflect.Ptr {
			base = base.Elem()
		}
		name := typeToTestName(base)

		ptrType := reflect.PointerTo(base)
		nilPtr := reflect.Zero(ptrType).Interface()

		cases[name+"nil"] = DeepCopyParams{
			Value: nilPtr,
		}
		cases[name+"value"] = DeepCopyParams{
			Value: random.Random(nilPtr),
		}
	}

	return cases
}

// DeepCopy provides a test function that tests a the `DeepCopy*` functions
// generated by `k8s.io/code-generator/cmd/deepcopy-gen` as part of the type
// system.
//
// The test function verifies that the copied value is equal to the original
// value but not the same reference. For simplicity, the function only requires
// a template to generate random non-zero test values that ensure covering all
// code paths. If this fails, you can vary the random seed, as well as the
// limits on the slice and map sizes (`size`) and string lengths (`len`).
//
// The following code shows a quick example of how to use this function in a
// tests:
//
// ```go
//
//	var deepCopyTestCases = DeepCopyTestCases(42, 3, 10,
//		   &MyStruct{}, (*MyStruct)(nil), ...)
//
//	func TestDeepCopy(t *testing.T) {
//		test.Map(t, deepCopyTestCases).Run(test.DeepCopy)
//	}
//
// ```
// *Note:* the test cases can also be generated inside the `TestDeepCopy`
// function.
func DeepCopy(t Test, p DeepCopyParams) {
	// Given
	value := p.Value

	// When
	result, err := deepCopy(value)
	if errors.Is(err, errNoDeepCopyMethod) {
		t.Fatalf("no deep copy method [%T]", value)
	}

	// Then
	rv := reflect.ValueOf(value)
	if rv.IsNil() {
		assert.Nil(t, result)

		return // Different pointer types cannot be equal.
	}

	// Wrap result in pointer if original value is pointer.
	if rv.Kind() == reflect.Ptr {
		// Create a new pointer to the result value.
		rvr := reflect.ValueOf(result)
		if rvr.Kind() != reflect.Ptr {
			presult := reflect.New(rvr.Type())
			presult.Elem().Set(rvr)
			result = presult.Interface()
		}

		// Only check not Same for two pointer types.
		assert.NotSame(t, value, result)
	}
	assert.Equal(t, value, result)
}

// deepCopy performs a deep copy of the given value using the either the
// `DeepCopyObject` or the `DeepCopy` method - in this order.
func deepCopy(value any) (any, error) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, errNoDeepCopyMethod
	}

	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		elem := rv.Type().Elem()
		zeroValue := reflect.Zero(elem)

		if method, ok := elem.MethodByName("DeepCopyObject"); ok {
			return zeroValue.Method(method.Index).
				Call(nil)[0].Interface(), nil
		}

		if method, ok := elem.MethodByName("DeepCopy"); ok {
			return zeroValue.Method(method.Index).
				Call(nil)[0].Interface(), nil
		}
	}

	method := rv.MethodByName("DeepCopyObject")
	if method.IsValid() {
		return method.Call(nil)[0].Interface(), nil
	}

	method = rv.MethodByName("DeepCopy")
	if method.IsValid() {
		return method.Call(nil)[0].Interface(), nil
	}

	return nil, errNoDeepCopyMethod
}

// errNoDeepCopyMethod is the error returned when no deep copy method is found.
var errNoDeepCopyMethod = errors.New("no deep copy method")
