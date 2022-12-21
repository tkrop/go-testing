package test

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	gosync "sync"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/sync"
	"github.com/tkrop/testing/utils/slices"
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

	// unknownName default unknown test case name.
	unknownName Name = "unknown"

	// Flag to run test by default sequential instead of parallel.
	Parallel = true
)

// Test is a minimal interface for abstracting methods on `testing.T` in a
// way that allows to setup an isolated test environment.
type Test interface {
	gomock.TestHelper
	require.TestingT
	Name() string
}

// Cleanuper defines an interface to add a custom mehtod that is called after
// the test execution to cleanup the test environment.
type Cleanuper interface {
	Cleanup(f func())
}

// Tester is a test isolation environment based on the `Test` abstraction. It
// can be used as a drop in replacement for `testing.T` in various libraries
// to check for expected test failures.
type Tester struct {
	Test
	sync.Synchronizer
	t        Test
	wg       sync.WaitGroup
	mu       gosync.Mutex
	failed   atomic.Bool
	cleanups []func()
	expect   Expect
}

// NewTester creates a new minimal test context based on the given `go-test`
// context.
func NewTester(t Test, expect Expect) *Tester {
	if tx, ok := t.(*Tester); ok {
		return &Tester{t: tx, wg: tx.wg, expect: expect}
	}
	return &Tester{t: t, expect: expect}
}

// Parallel delegates request to `testing.T.Parallel()`.
func (t *Tester) Parallel() {
	if t, ok := t.t.(*testing.T); ok {
		t.Parallel()
	}
}

// Cleanup is a function called to setup test cleanup after execution. This
// method is allowing `gomock` to register its `finish` method that reports the
// missing mock calls.
func (t *Tester) Cleanup(cleanup func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cleanups = append(t.cleanups, cleanup)
}

// WaitGroup adds wait group to unlock in case of a failure.
func (t *Tester) WaitGroup(wg sync.WaitGroup) {
	t.wg = wg
}

// Name delegates the request to the parent test context.
func (t *Tester) Name() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.t.Name()
}

// Helper delegates request to the parent test context.
func (t *Tester) Helper() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.t.Helper()
}

// Unlock unlocks the wait group of the test by consuming the wait group
// counter completely.
func (t *Tester) Unlock() {
	if t.wg != nil {
		t.wg.Add(math.MinInt)
	}
}

// FailNow delegates failure handling delayed to the parent test context.
func (t *Tester) FailNow() {
	t.Helper()
	t.failure(func() {
		t.t.FailNow()
	}, true)
}

// Errorf delegated the failure handling to the parent test context.
func (t *Tester) Errorf(format string, args ...any) {
	t.Helper()
	t.failure(func() {
		t.t.Errorf(format, args...)
	}, false)
}

// Fatalf delegated the failure handling to the parent test context.
func (t *Tester) Fatalf(format string, args ...any) {
	t.Helper()
	t.failure(func() {
		t.t.Fatalf(format, args...)
	}, true)
}

// failure handles all failures by setting the failure state stopping the test
// and reporting the failure to the parent context - if success was expected.
// The failure is always scheduled to be reported during cleanup by the parent
// test go-routine - to open the test framework for spawning new go-routines
// (not possible by now).
func (t *Tester) failure(fail func(), exit bool) {
	t.Helper()
	t.failed.Store(true)
	if t.expect == Success {
		t.Cleanup(fail)
	}
	if exit {
		defer t.Unlock()
		runtime.Goexit()
	}
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

	// register clean up handlers with test context.
	if c, ok := t.t.(Cleanuper); ok {
		c.Cleanup(func() {
			t.Helper()
			t.cleanup()
		})
	}

	// register finish as first clean up handler.
	t.Cleanup(func() {
		t.Helper()
		t.finish()
	})

	// execute test function.
	wg := sync.NewWaitGroup()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer t.recover()
		test(t)
	}()
	wg.Wait()

	return t
}

// recover recovers from panics and generate test failure.
func (t *Tester) recover() {
	t.Helper()
	if err := recover(); err != nil {
		t.Fatalf("panic: %v", err)
	}
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

// Runner is a generic test runner interface.
type Runner[P any] interface {
	// Run runs the test parameter sets (by default) parallel.
	Run(call func(t Test, param P))
	// RunSeq runs the test parameter sets in a sequence.
	RunSeq(call func(t Test, param P))
}

// runner is a generic parameterized test runner struct.
type runner[P any] struct {
	t      *testing.T
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
		params: params,
	}
}

// Map creates a new parallel test runner with given test parameter sets
// provided as a test case name to parameter sets mapping.
func Map[P any](t *testing.T, params map[string]P) Runner[P] {
	t.Helper()

	return New[P](t, params)
}

// Slice creates a new parallel test runner with given test parameter sets
// provided as a slice. The test runner is looking into the parameter set to
// find a suitable test case name.
func Slice[P any](t *testing.T, params []P) Runner[P] {
	t.Helper()

	return New[P](t, params)
}

// Run runs the test parameter sets (by default) parallel.
func (r *runner[P]) Run(call func(t Test, param P)) {
	r.run(call, Parallel)
}

// RunSeq runs the test parameter sets in a sequence.
func (r *runner[P]) RunSeq(call func(t Test, param P)) {
	r.run(call, false)
}

// Run runs the test parameter sets either parallel or in sequence.
func (r *runner[P]) run(call func(t Test, param P), parallel bool) {
	switch params := r.params.(type) {
	case map[string]P:
		if parallel {
			r.t.Parallel()
		}

		for name, param := range params {
			name, param := name, param
			r.t.Run(name, run(r.expect(param), func(t Test) {
				// Helpful for debugging to see the test case.
				require.NotEmpty(t, name)

				call(t, param)
			}, parallel))
		}

	case []P:
		if parallel {
			r.t.Parallel()
		}

		for index, param := range params {
			index, param := index, param
			name := fmt.Sprintf("%s[%d]", r.name(param), index)
			r.t.Run(name, run(r.expect(param), func(t Test) {
				// Helpful for debugging to see the test case.
				require.NotEmpty(t, name)

				call(t, param)
			}, parallel))
		}
	case P:
		name := string(r.name(params))
		if name == string(unknownName) {
			run(r.expect(params), func(t Test) {
				// Helpful for debugging to see the test case.
				require.NotEmpty(t, name)

				call(t, params)
			}, parallel)(r.t)
		} else {
			r.t.Run(name, run(r.expect(params), func(t Test) {
				// Helpful for debugging to see the test case.
				require.NotEmpty(t, name)

				call(t, params)
			}, parallel))
		}
	default:
		panic(fmt.Errorf("unknown parameter type:  %v",
			reflect.ValueOf(r.params).Type()))
	}
}

// name resolves the test case name from the parameter set.
func (r *runner[P]) name(param P) Name {
	name, ok := extract(param, unknownName, "name").(Name)
	if ok && name != "" {
		return name
	}
	return unknownName
}

// expect resolves the test case expectation from the parameter set.
func (r *runner[P]) expect(param P) Expect {
	if expect, ok := extract(param, Success, "expect").(Expect); ok {
		return expect
	}
	return Success
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
	return run(expect, test, false)
}

// Run creates an isolated parallel or sequential test environment running the
// given test function with given expectation. When executed via `t.Run()` it
// checks whether the result is matching the expectation.
func run(expect Expect, test func(Test), parallel bool) func(*testing.T) {
	return func(t *testing.T) {
		NewTester(t, expect).Run(test, parallel)
	}
}

// InRun creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expectation.
func InRun(expect Expect, test func(Test)) func(Test) {
	return func(t Test) {
		NewTester(t, expect).Run(test, false)
	}
}
