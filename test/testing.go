package test

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	gosync "sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/internal/slices"
	"github.com/tkrop/go-testing/internal/sync"
)

type (
	// Expect the expectation whether a test will succeed or fail.
	Expect bool
	// Name represents a test case name.
	Name string
)

// Constants to express test expectations.
const (
	// Success used to express that a test is supposed to succeed.
	Success Expect = true
	// Failure used to express that a test is supposed to fail.
	Failure Expect = false

	// unknown default unknown test case name.
	unknown Name = "unknown"

	// Flag to run test by default sequential instead of parallel.
	Parallel = true
)

// TODO: consider following convenience methods:
//
// // Result is a convenience method that returns the first argument ans swollows
// // all others assuming that the first argument contains the important result to
// // focus the test at.
// func Result[T any](result T, swollowed any) T {
// 	return result
// }

// // Check is a convenience method that returns the second argument and swollows
// // the first used to focus a test on the second.
// func Check[T any](swollowed any, check T) T {
// 	return check
// }

// // NoError is a convenience method to check whether the second error argument
// // is providing and actual error while extracting the first argument only. If
// // the error argument is an error, the method panics providing the error.
// func NoError[T any](result T, err error) T {
// 	if err != nil {
// 		panic(err)
// 	}
// 	return result
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

// TestName returns the normalized test case name for the given name and given
// parameter set. If the name is empty, the name is resolved from the parameter
// set using the `name` parameter. The resolved name is normalized before being
// returned.
func TestName[P any](name string, param P) string {
	if name != "" {
		return strings.ReplaceAll(name, " ", "-")
		// TODO: replace reflect.FindArgOf with Find - needs better structs support!
		// } else if name := Find(param, unknown, "name"); name != "" {
		// 	return strings.ReplaceAll(string(name), " ", "-")
	}
	found := reflect.FindArgOf(param, unknown, "name")
	if name, ok := found.(Name); ok && name != "" {
		return strings.ReplaceAll(string(name), " ", "-")
	}
	return string(unknown)
}

// TestExpect resolves the test case expectation from the parameter set. If no
// expectation is found, the default expectation `Success` is returned.
func TestExpect[P any](param P) Expect {
	// TODO: replace reflect.FindArgOf with Find - needs better struct support!
	// return Find(param, Success, "expect")
	expect := reflect.FindArgOf(param, Success, "expect")
	if expect, ok := expect.(Expect); ok {
		return expect
	}
	return Success
}

// Reporter is a minimal interface for abstracting test report methods that are
// needed to setup an isolated test environment for GoMock and Testify.
type Reporter interface {
	// Panic reports a panic.
	Panic(arg any)
	// Errorf reports a failure messages when a test is supposed to continue.
	Errorf(format string, args ...any)
	// Fatalf reports a fatal failure message that immediate aborts of the test
	// execution.
	Fatalf(format string, args ...any)
	// FailNow reports fatal failure notifications without log output that
	// aborts test execution immediately.
	FailNow()
}

// Test is a minimal interface for abstracting test methods that are needed to
// setup an isolated test environment for GoMock and Testify.
type Test interface {
	// Name provides the test name.
	Name() string
	// Helper declares a test helper function.
	Helper()
	// Parallel declares that the test is to be run in parallel with (and only
	// with) other parallel tests.
	Parallel()
	// TempDir creates a new temporary directory for the test.
	TempDir() string
	// Errorf handles a failure messages when a test is supposed to continue.
	Errorf(format string, args ...any)
	// Fatalf handles a fatal failure message that immediate aborts of the test
	// execution.
	Fatalf(format string, args ...any)
	// FailNow handles fatal failure notifications without log output that
	// aborts test execution immediately.
	FailNow()
	// Setenv sets an environment variable for the test.
	Setenv(key, value string)
}

// Cleanuper defines an interface to add a custom mehtod that is called after
// the test execution to cleanup the test environment.
type Cleanuper interface {
	Cleanup(cleanup func())
}

// Tester is a test isolation environment based on the `Test` abstraction. It
// can be used as a drop in replacement for `testing.T` in various libraries
// to check for expected test failures.
type Tester struct {
	sync.Synchronizer
	t        Test
	wg       sync.WaitGroup
	mu       gosync.Mutex
	failed   atomic.Bool
	reporter Reporter
	cleanups []func()
	expect   Expect
}

// NewTester creates a new minimal test context based on the given `go-test`
// context.
func NewTester(t Test, expect Expect) *Tester {
	if tx, ok := t.(*Tester); ok {
		return (&Tester{t: tx, wg: tx.wg, expect: expect})
	}
	return (&Tester{t: t, expect: expect})
}

// WaitGroup adds wait group to unlock in case of a failure.
//
//revive:disable-next-line:waitgroup-by-value // own wrapper interface
func (t *Tester) WaitGroup(wg sync.WaitGroup) {
	t.wg = wg
}

// Reporter sets up a test failure reporter. This can be used to validate the
// reported failures in a test environment.
func (t *Tester) Reporter(reporter Reporter) {
	t.reporter = reporter
}

// Cleanup is a function called to setup test cleanup after execution. This
// method is allowing `gomock` to register its `finish` method that reports the
// missing mock calls.
func (t *Tester) Cleanup(cleanup func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cleanups = append(t.cleanups, cleanup)
}

// Name delegates the request to the parent test context.
func (t *Tester) Name() string {
	return t.t.Name()
}

// Helper delegates request to the parent test context.
func (t *Tester) Helper() {
	t.t.Helper()
}

// Parallel robustly delegates request to the parent context. It can be called
// multiple times, since it is swallowing the panic that is raised when calling
// `t.Parallel()` multiple times.
func (t *Tester) Parallel() {
	defer t.recoverParallel()
	t.t.Parallel()
}

// Setenv delegates request to the parent context, if it is of type
// `*testing.T`. Else it is swallowing the request silently.
func (t *Tester) Setenv(key, value string) {
	t.t.Setenv(key, value)
}

// TempDir delegates the request to the parent test context.
func (t *Tester) TempDir() string {
	return t.t.TempDir()
}

// Errorf handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context. Else it delegates the request to the test reporter if available.
func (t *Tester) Errorf(format string, args ...any) {
	t.Helper()
	t.failed.Store(true)
	if t.expect == Success {
		t.t.Errorf(format, args...)
	} else if t.reporter != nil {
		t.reporter.Errorf(format, args...)
	}
}

// Fatalf handles a fatal failure message that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context. Else it delegates the request to the test
// reporter if available.
func (t *Tester) Fatalf(format string, args ...any) {
	t.Helper()
	t.failed.Store(true)
	defer t.unlock()
	if t.expect == Success {
		t.t.Fatalf(format, args...)
	} else if t.reporter != nil {
		t.reporter.Fatalf(format, args...)
	}
	runtime.Goexit()
}

// FailNow handles fatal failure notifications without log output that aborts
// test execution immediately. On an expected success, it the failure handling
// is also delegated to the parent test context. Else it delegates the request
// to the test reporter if available.
func (t *Tester) FailNow() {
	t.Helper()
	t.failed.Store(true)
	defer t.unlock()
	if t.expect == Success {
		t.t.FailNow()
	} else if t.reporter != nil {
		t.reporter.FailNow()
	}
	runtime.Goexit()
}

// Offset fr original stack in case of panic handling.
const panicOriginStackOffset = 10

// Panic handles failure notifications of panics that also abort the test
// execution immediately.
func (t *Tester) Panic(arg any) {
	t.Helper()
	t.failed.Store(true)
	defer t.unlock()
	if t.expect == Success {
		stack := strings.SplitN(string(debug.Stack()), "\n",
			panicOriginStackOffset)
		t.Fatalf("panic: %v\n%s\n%s", arg, stack[0],
			stack[panicOriginStackOffset-1])
	} else if t.reporter != nil {
		t.reporter.Panic(arg)
	}
	runtime.Goexit()
}

// Run executes the test function in a safe detached environment and check
// the failure state after the test function has finished. If the test result
// is not according to expectation, a failure is created in the parent test
// context.
func (t *Tester) Run(test func(Test), parallel bool) Test {
	t.Helper()
	if parallel {
		t.Parallel()
	}

	// register cleanup handlers.
	t.register()

	// execute test function.
	wg := sync.NewWaitGroup()
	wg.Add(1)
	go func() {
		t.Helper()
		defer wg.Done()
		defer t.recover()
		test(t)
	}()
	wg.Wait()

	return t
}

// register registers the clean up handlers with the parent test context.
func (t *Tester) register() {
	t.Helper()

	if c, ok := t.t.(Cleanuper); ok {
		c.Cleanup(func() {
			t.Helper()
			t.cleanup()
		})
	}

	t.Cleanup(func() {
		t.Helper()
		t.finish()
	})
}

// cleanup runs the cleanup methods registered on the isolated test environment.
func (t *Tester) cleanup() {
	t.mu.Lock()
	cleanups := slices.Reverse(t.cleanups)
	t.mu.Unlock()

	for _, cleanup := range cleanups {
		cleanup()
	}
}

// finish evaluates the final result of the test function in relation to the
// provided expectation.
func (t *Tester) finish() {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.expect {
	case Success:
		if t.failed.Load() {
			t.t.Errorf("Expected test to succeed but it failed: %s", t.t.Name())
		}
	case Failure:
		if !t.failed.Load() {
			t.t.Errorf("Expected test to fail but it succeeded: %s", t.t.Name())
		}
	}
}

// recover recovers from panics and generate test failure.
func (t *Tester) recover() {
	t.Helper()

	//revive:disable-next-line:defer // only used inside a deferred call.
	if arg := recover(); arg != nil {
		t.Panic(arg)
	}
}

// recoverParallel recovers from panics when calling `t.Parallel()` multiple
// times.
func (t *Tester) recoverParallel() {
	t.Helper()

	//revive:disable-next-line:defer // only used inside a deferred call.
	if v := recover(); v != nil &&
		v != "testing: t.Parallel called multiple times" {
		t.Panic(v)
	}
}

// unlock unlocks the wait group of the test by consuming the wait group
// counter completely.
func (t *Tester) unlock() {
	if t.wg != nil {
		t.wg.Add(math.MinInt)
	}
}

// Runner is a generic test runner interface.
type Runner[P any] interface {
	// Filter sets up a filter for the test cases using the given pattern and
	// match flag. The pattern is a regular expression that is matched against
	// the test case name. The match flag is used to include or exclude the
	// test cases that match the pattern.
	Filter(pattern string, match bool) Runner[P]
	// Run runs all test parameter sets in parallel. If the test parameter sets
	// are provided as a map, the test case name is used as the test name. If
	// the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	Run(call func(t Test, param P)) Runner[P]
	// RunSeq runs the test parameter sets in a sequence. If the test parameter
	// sets are provided as a map, the test case name is used as the test name.
	// If the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	RunSeq(call func(t Test, param P)) Runner[P]
	// Cleanup register a function to be called to cleanup after all tests have
	// finished to remove the shared resources.
	Cleanup(call func())
}

// runner is a generic parameterized test runner struct.
type runner[P any] struct {
	t      *testing.T
	wg     sync.WaitGroup
	filter func(string) bool
	params any
}

// New creates a new parallel test runner with given parameter sets, i.e. a
// single test parameter set, a slice of test parameter sets, or a test case
// name to test parameter set map. If necessary, the test runner is looking
// into the parameter set for a suitable test case name.
func New[P any](t *testing.T, params any) Runner[P] {
	t.Helper()

	return &runner[P]{
		t:      t,
		wg:     sync.NewWaitGroup(),
		params: params,
	}
}

// Map creates a new parallel test runner with given test parameter sets
// provided as a test case name to parameter sets mapping.
func Map[P any](t *testing.T, params ...map[string]P) Runner[P] {
	t.Helper()

	return New[P](t, maps.Add(maps.Copy(params[0]), params[1:]...))
}

// Slice creates a new parallel test runner with given test parameter sets
// provided as a slice. The test runner is looking into the parameter set to
// find a suitable test case name.
func Slice[P any](t *testing.T, params []P) Runner[P] {
	t.Helper()

	return New[P](t, params)
}

// Filter filters the test cases by the given pattern and match flag. The
// pattern is a regular expression that is matched against the test case name.
// The match flag is used to include or exclude the test cases that match the
// pattern.
func (r *runner[P]) Filter(
	pattern string, match bool,
) Runner[P] {
	regexp := regexp.MustCompile(pattern)
	r.filter = func(name string) bool {
		return regexp.MatchString(name) == match
	}
	return r
}

// Run runs the test parameter sets (by default) parallel.
func (r *runner[P]) Run(call func(t Test, param P)) Runner[P] {
	return r.run(call, Parallel)
}

// RunSeq runs the test parameter sets in a sequence.
func (r *runner[P]) RunSeq(call func(t Test, param P)) Runner[P] {
	return r.run(call, !Parallel)
}

// Cleanup register a function to be called for cleanup after all tests have
// been finished.
func (r *runner[P]) Cleanup(call func()) {
	r.t.Cleanup(func() {
		r.t.Helper()
		r.wg.Wait()
		call()
	})
}

// Parallel ensures that the test runner runs the test parameter sets in
// parallel.
func (r *runner[P]) parallel(parallel bool) {
	if parallel {
		defer r.recoverParallel()
		r.t.Parallel()
	}
}

// RecoverParallel recovers from panics when calling `t.Parallel()` multiple
// times.
func (*runner[P]) recoverParallel() {
	//revive:disable-next-line:defer // only used inside a deferred call.
	if v := recover(); v != nil &&
		v != "testing: t.Parallel called multiple times" {
		panic(v)
	}
}

// Run runs the test parameter sets either parallel or in sequence.
func (r *runner[P]) run(
	call func(t Test, param P), parallel bool,
) Runner[P] {
	switch params := r.params.(type) {
	case map[string]P:
		r.parallel(parallel)
		for name, param := range params {
			name := TestName(name, param)
			if r.filter != nil && !r.filter(name) {
				continue
			}
			r.wg.Add(1)
			r.t.Run(name, r.wrap(name, param, call, parallel))
		}

	case []P:
		r.parallel(parallel)
		for index, param := range params {
			name := TestName("", param) + "[" + strconv.Itoa(index) + "]"
			if r.filter != nil && !r.filter(name) {
				continue
			}
			r.wg.Add(1)
			r.t.Run(name, r.wrap(name, param, call, parallel))
		}

	case P:
		name := TestName("", params)
		if r.filter != nil && !r.filter(name) {
			return r
		}
		r.wg.Add(1)
		if name != string(unknown) {
			r.t.Run(name, r.wrap(name, params, call, parallel))
		} else {
			r.wrap(name, params, call, parallel)(r.t)
		}

	default:
		panic(NewErrUnknownParameterType(r.params))
	}
	return r
}

// Wrap creates the test wrapper method executing the test.
func (r *runner[P]) wrap(
	name string, param P, call func(t Test, param P), parallel bool,
) func(*testing.T) {
	return run(TestExpect(param), func(t Test) {
		t.Helper()

		// Helpful for debugging to see the test case.
		require.NotEmpty(t, name)

		defer r.wg.Done()
		call(t, param)
	}, parallel)
}

// Run creates an isolated (by default) parallel test environment running the
// given test function with given expectation. When executed via `t.Run()` it
// checks whether the result is matching the expectation.
func Run(expect Expect, test func(Test)) func(*testing.T) {
	return run(expect, test, Parallel)
}

// RunSeq creates an isolated, test environment for the given test function
// with given expectation. When executed via `t.Run()` it checks whether the
// result is matching the expectation.
func RunSeq(expect Expect, test func(Test)) func(*testing.T) {
	return run(expect, test, !Parallel)
}

// Run creates an isolated parallel or sequential test environment running the
// given test function with given expectation. When executed via `t.Run()` it
// checks whether the result is matching the expectation.
func run(expect Expect, test func(Test), parallel bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		NewTester(t, expect).Run(test, parallel)
	}
}

// InRun creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expectation.
func InRun(expect Expect, test func(Test)) func(Test) {
	return func(t Test) {
		t.Helper()

		NewTester(t, expect).Run(test, !Parallel)
	}
}

// ErrUnknownParameterType is an error for unknown parameter types.
var ErrUnknownParameterType = errors.New("unknown parameter type")

// NewErrUnknownParameterType creates a new unknown parameter type error.
func NewErrUnknownParameterType(value any) error {
	return fmt.Errorf("%w [type: %v]",
		ErrUnknownParameterType, reflect.ValueOf(value).Type())
}
