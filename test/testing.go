package test

import (
	"fmt"
	"math"
	"runtime"
	gosync "sync"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/internal/slices"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/sync"
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

// Reporter is a minimal inferface for abstracting test report methods that are
// needed to setup an isolated test environment for GoMock and Testify.
type Reporter interface {
	Panic(arg any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	FailNow()
}

// Test is a minimal interface for abstracting test methods that are needed to
// setup an isolated test environment for GoMock and Testify.
type Test interface {
	Helper()
	Name() string
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	FailNow()
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

// Parallel delegates request to `testing.T.Parallel()`.
func (t *Tester) Parallel() {
	if t, ok := t.t.(*testing.T); ok {
		t.Parallel()
	}
}

// WaitGroup adds wait group to unlock in case of a failure.
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

// Errorf handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context.
func (t *Tester) Errorf(format string, args ...any) {
	t.Helper()
	t.failed.Store(true)
	if t.expect == Success {
		t.t.Errorf(format, args...)
	} else if t.reporter != nil {
		t.reporter.Errorf(format, args...)
	}
}

// Fatalf handles a fatal failure messge that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context.
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
// is also delegated to the parent test context.
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

// Panic handles failure notifications of panics that also abort the test
// execution immediately.
func (t *Tester) Panic(arg any) {
	t.Helper()
	t.failed.Store(true)
	defer t.unlock()
	if t.expect == Success {
		t.Fatalf("panic: %v", arg)
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
	if arg := recover(); arg != nil {
		t.Panic(arg)
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
			r.t.Run(name, r.wrap(name, param, call, parallel))
		}

	case []P:
		if parallel {
			r.t.Parallel()
		}

		for index, param := range params {
			index, param := index, param
			name := fmt.Sprintf("%s[%d]", r.name(param), index)
			r.t.Run(name, r.wrap(name, param, call, parallel))
		}
	case P:
		name := string(r.name(params))
		if name != string(unknownName) {
			r.t.Run(name, r.wrap(name, params, call, parallel))
		} else {
			r.wrap(name, params, call, parallel)(r.t)
		}
	default:
		panic(fmt.Errorf("unknown parameter type:  %v",
			reflect.ValueOf(r.params).Type()))
	}
}

// wrap creates the test wrapper method executing the test.
func (r *runner[P]) wrap(
	name string, param P, call func(t Test, param P), parallel bool,
) func(*testing.T) {
	return run(r.expect(param), func(t Test) {
		// Helpful for debugging to see the test case.
		require.NotEmpty(t, name)

		call(t, param)
	}, parallel)
}

// name resolves the test case name from the parameter set.
func (r *runner[P]) name(param P) Name {
	name, ok := reflect.FindArgOf(param, unknownName, "name").(Name)
	if ok && name != "" {
		return name
	}
	return unknownName
}

// expect resolves the test case expectation from the parameter set.
func (r *runner[P]) expect(param P) Expect {
	if expect, ok := reflect.FindArgOf(param, Success, "expect").(Expect); ok {
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

		NewTester(t, expect).Run(test, false)
	}
}

// Validator a test failure validator based on the thest reporter interface.
type Validator struct {
	ctrl     *gomock.Controller
	recorder *Recorder
}

// Recorder a test failure validator recorder.
type Recorder struct {
	validator *Validator
}

func NewValidator(ctrl *gomock.Controller) *Validator {
	validator := &Validator{ctrl: ctrl}
	validator.recorder = &Recorder{validator: validator}
	if t, ok := ctrl.T.(*Tester); ok {
		ctrl.T = NewTester(t.t, Success)
		t.Reporter(validator)
	}
	return validator
}

func (v *Validator) EXPECT() *Recorder {
	return v.recorder
}

// Errorf receive expected method call to `Errorf`.
func (v *Validator) Errorf(format string, args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Errorf", append([]any{format}, args...)...)
}

// Errorf indicate an expected method call to `Errorf`.
func (r *Recorder) Errorf(format string, args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Errorf",
		reflect.TypeOf((*Validator)(nil).Errorf), append([]any{format}, args...)...)
}

// Errorf creates a validation method call setup for `Errorf`.
func Errorf(format string, args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Errorf(format, args...).Times(mocks.Times(1)).
			Do(mocks.GetVarDone(2))
	}
}

// Fatalf receive expected method call to `Fatalf`.
func (v *Validator) Fatalf(format string, args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Fatalf", append([]any{format}, args...)...)
}

// Fatalf indicate an expected method call to `Fatalf`.
func (r *Recorder) Fatalf(format string, args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Fatalf",
		reflect.TypeOf((*Validator)(nil).Fatalf), append([]any{format}, args...)...)
}

// Fatalf creates a validation method call setup for `Fatalf`.
func Fatalf(format string, args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Fatalf(format, args...).Times(mocks.Times(1)).
			Do(mocks.GetVarDone(2))
	}
}

// FailNow receive expected method call to `FailNow`.
func (v *Validator) FailNow() {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "FailNow")
}

// FailNow indicate an expected method call to `FailNow`.
func (r *Recorder) FailNow() *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "FailNow",
		reflect.TypeOf((*Validator)(nil).FailNow))
}

// FailNow creates a validation method call setup for `FailNow`.
func FailNow() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			FailNow().Times(mocks.Times(1)).
			Do(mocks.GetDone(0))
	}
}

// Panic receive expected method call indicating a panic.
func (v *Validator) Panic(arg any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Panic", []any{arg}...)
}

// Panic indicate an expected method call from panic.
func (r *Recorder) Panic(arg any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Panic",
		reflect.TypeOf((*Validator)(nil).Panic), []any{arg}...)
}

// Panic creates a validation method call setup for a panic.
func Panic(arg any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Panic(arg).Times(mocks.Times(1)).
			Do(mocks.GetDone(1))
	}
}
