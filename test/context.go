package test

import (
	"math"
	"regexp"
	"runtime"
	"runtime/debug"
	gosync "sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/internal/sync"
)

// Types and constants for the test context and test runner.
type (
	// Expect the expectation whether a test will succeed or fail.
	Expect bool
	// Mode defines the test execution mode. Currently mode only supports
	// either parallel or sequential.
	Mode int
)

// Constants to express test expectations.
const (
	// Success used to express that a test is supposed to succeed.
	Success Expect = true
	// Failure used to express that a test is supposed to fail.
	Failure Expect = false
)

// Constants to express test execution modes.
const (
	// Sequential is the test execution mode to run tests in sequence.
	Sequential Mode = 0x0
	// Parallel is the test execution mode to run tests in parallel.
	Parallel Mode = 0x1
)

// Test is a minimal interface for abstracting test methods that are needed to
// setup an isolated test environment for GoMock and Testify.
type Test interface { //nolint:interfacebloat // Minimal interface.
	// Embeds the basic test reporter interface.
	Reporter
	// Name provides the test name.
	Name() string
	// Helper declares a test helper function.
	Helper()
	// Parallel declares that the test is to be run in parallel with (and only
	// with) other parallel tests.
	Parallel()
	// TempDir creates a new temporary directory for the test.
	TempDir() string
	// Setenv sets an environment variable for the test.
	Setenv(key, value string)
	// Deadline returns the deadline of the test and a flag indicating whether
	// the deadline is set.
	Deadline() (deadline time.Time, ok bool)
	// Skip is a helper method to skip the test.
	Skip(args ...any)
	// Skipf is a helper method to skip the test with a formatted message.
	Skipf(format string, args ...any)
	// SkipNow is a helper method to skip the test immediately.
	SkipNow()
	// Skipped reports whether the test has been skipped.
	Skipped() bool
	// Failed reports whether the test has failed.
	Failed() bool
	// Cleanup is a function called to setup test cleanup after execution.
	Cleanup(cleanup func())
}

type Tester interface {
	// Embeds the minimal test interface.
	Test
	// Embeds the test failure reporter interface.
	Panicer
	// Mode sets up the test execution mode, either `Parallel` or `Sequential`.
	//
	// **Note:** `Parallel` only affects the current test context before the
	// test execution is started, i.e. before `Run` is called. After the test
	// execution is started, calling this method only affects sub-tests.
	Mode(mode Mode) Tester
	// Expect sets up a different expected test outcome, i.e. `test.Success` or
	// `test.Failure`. Can be called multiple times, but the last call wins.
	Expect(expect Expect) Tester
	// Timeout sets up an individual timeout for the test. The method does not
	// affect the global test timeout or a pending parent timeout that may abort
	// the test, if the given duration is exceeding the timeout.
	//
	// A negative or zero duration is ignored and will not change the timeout. If
	// this method is called multiple times, the last call wins.
	Timeout(timeout time.Duration) Tester
	// StopEarly stops the test by the given duration ahead of the individual or
	// global test deadline, to ensure that a cleanup function has sufficient time
	// to finish before the deadline is exceeded. The method is not able to extend
	// the test deadline.
	//
	// A negative or zero duration is ignored. **Warning:** calling this method
	// multiple times will also reduce the test time step by step.
	StopEarly(time time.Duration) Tester
	// WaitGroup adds wait group to unlock in case of a failure.
	//
	//revive:disable-next-line:waitgroup-by-value // own wrapper interface
	WaitGroup(wg sync.WaitGroup)
	// Reporter sets up a test failure reporter. This can be used to validate the
	// reported failures in a test environment.
	Reporter(reporter Reporter)
	// Run executes the test function in a safe detached environment and checks
	// the failure state after the test function has finished. If the expectation
	// is not met, a failure is created in the parent test context.
	Run(name string, call Func)
}

// Cleanuper defines an interface to add a custom method that is called after
// the test execution to cleanup the test environment.
type Cleanuper interface {
	Cleanup(cleanup func())
}

// Func defines the common test function signature.
type Func func(Test)

// Run creates an isolated (by default) parallel test context running the given
// test function with success expectation. If the expectation is not met, a test
// failure is created in the parent test context.
//
// **Note:** You can still call `Expect(Failure)` to change the expected test
// outcome. You may also call `Parallel()`, since the context is swallowing the
// panic that is raised when calling `Parallel()` multiple times.
func Run(test func(Tester)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		New(t).Run("", func(t Test) {
			t.Helper()

			test(Cast[Tester](t))
		})
	}
}

// RunSeq creates an isolated, test context for the given test function with
// default success expectation. If the expectation is not met, a test failure
// is created in the parent test context.
//
// **Note:** You can still call `Expect(Failure)` to change the expected test
// outcome. You may also call `Parallel()` to parallelize the test execution.
// Repeated calls to `Parallel()` are also allowed, since the context is
// swallowing the raised panic.
func RunSeq(test func(Tester)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		New(t).Mode(Sequential).Run("", func(t Test) {
			t.Helper()

			test(Cast[Tester](t))
		})
	}
}

// Wrap wraps a test function into an isolated, sequential test context with
// given test expectation. If the expectation is not met, the wrapper creates a
// failure in the parent test context.
//
// **Note:** You can call `Parallel()` to parallelize the test execution.
// If you cast the basic `Test` interface to a `Tester`, you may also call
// `Expect(Success|Failure)`, `Timeout()`, and `StopEarly()` to change the
// test behavior and expectations.
func Wrap(expect Expect, test Func) Func {
	return func(t Test) {
		t.Helper()

		New(t).Mode(Sequential).Expect(expect).Run("", test)
	}
}

// Context is a test isolation environment based on the `Test` abstraction. It
// can be used as a drop in replacement for `testing.T` in various libraries
// to check for expected test failures.
type Context struct {
	sync.Synchronizer
	t        Test
	wg       sync.WaitGroup
	mu       gosync.Mutex
	failed   atomic.Bool
	deadline time.Time
	reporter Reporter
	cleanups []func()
	expect   Expect
	mode     Mode
}

// New creates a new minimal isolated test context (`Tester`) based on the
// provided test context, executed in a safe detached, parallel environment
// expecting a successful test execution. The test context is delegating all
// method calls to the parent test context, which is used to propagate test
// results.
//
// If the provided test context is already of type `*Context`, the new context
// is created based on the existing state to simplify consistent nested context
// creation with same deadline, same wait group, and same expectations.
//
// **Note:** even though the test context is created with parallelization and
// successful execution expectation, this expectation can be changed by calling
// `Mode(Sequential)` or `Expect(Failure)` to change the behavior. The test
// function can also simply call `Parallel()`, since the context is silently
// swallowing the panic raised when calling `Parallel()` multiple times.
func New(t Test) Tester {
	if c, ok := t.(*Context); ok {
		return &Context{
			t: c, wg: c.wg,
			deadline: c.deadline,
			expect:   c.expect,
			mode:     c.mode,
		}
	}

	return &Context{
		t: t,
		deadline: func(t Test) time.Time {
			defer func() { _ = recover() }()
			deadline, _ := t.Deadline()
			return deadline
		}(t),
		expect: true,
		mode:   Parallel,
	}
}

// Mode sets up the test execution mode, either `Parallel` or `Sequential`.
//
// **Note:** `Parallel` only affects the current test context before the test
// execution is started, i.e. before `Run` is called. After the test execution
// is started, calling this method only affects sub-tests.
func (c *Context) Mode(mode Mode) Tester {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.mode = mode

	return c
}

// Expect sets up a different expected test outcome, i.e. `test.Success` or
// `test.Failure`. Can be called multiple times, but the last call wins.
func (c *Context) Expect(expect Expect) Tester {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.expect = expect

	return c
}

// Timeout sets up an individual timeout for the test. The method does not
// affect the global test timeout or a pending parent timeout that may abort
// the test, if the given duration is exceeding the timeout.
//
// A negative or zero duration is ignored and will not change the timeout. If
// this method is called multiple times, the last call wins.
func (c *Context) Timeout(timeout time.Duration) Tester {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	if timeout > 0 {
		c.deadline = time.Now().Add(timeout)
	}

	return c
}

// StopEarly stops the test by the given duration ahead of the individual or
// global test deadline, to ensure that a cleanup function has sufficient time
// to finish before the deadline is exceeded. The method is not able to extend
// the test deadline.
//
// A negative or zero duration is ignored. **Warning:** calling this method
// multiple times will also reduce the test time step by step.
func (c *Context) StopEarly(time time.Duration) Tester {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.deadline.IsZero() && time > 0 {
		c.deadline = c.deadline.Add(-time)
	}

	return c
}

// WaitGroup adds wait group to unlock in case of a failure.
//
//revive:disable-next-line:waitgroup-by-value // own wrapper interface
func (c *Context) WaitGroup(wg sync.WaitGroup) {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.wg = wg
}

// Done decrements the wait group of the test by one.
func (c *Context) Done() {
	c.t.Helper()

	if c.wg != nil {
		c.wg.Done()
	}
}

// Reporter sets up a test failure reporter. This can be used to validate the
// reported failures in a test environment.
func (c *Context) Reporter(reporter Reporter) {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.reporter = reporter
}

// Cleanup is a function called to setup test cleanup after execution. This
// method is allowing `gomock` to register its `finish` method that reports the
// missing mock calls.
func (c *Context) Cleanup(cleanup func()) {
	c.t.Helper()
	if cleanup == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanups = append(c.cleanups, cleanup)
}

// Name delegates the request to the parent test context.
func (c *Context) Name() string {
	c.t.Helper()

	return c.t.Name()
}

// Helper delegates request to the parent test context.
func (c *Context) Helper() {
	c.t.Helper()
}

// Parallel robustly delegates request to the parent context. It can be called
// multiple times, since it is swallowing the panic that is raised when calling
// `t.Parallel()` multiple times.
func (c *Context) Parallel() {
	c.t.Helper()

	defer func() {
		if err := recover(); err != nil &&
			err != "testing: t.Parallel called multiple times" {
			c.Panic(err)
		}
	}()

	c.t.Parallel()
}

// TempDir delegates the request to the parent test context.
func (c *Context) TempDir() string {
	c.t.Helper()
	return c.t.TempDir()
}

// Setenv delegates request to the parent context, if it is of type
// `*testing.T`. Else it is swallowing the request silently.
func (c *Context) Setenv(key, value string) {
	c.t.Helper()

	c.t.Setenv(key, value)
}

// Deadline delegates request to the parent context. It returns the deadline of
// the test and a flag indicating whether the deadline is set.
func (c *Context) Deadline() (time.Time, bool) {
	c.t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.deadline.IsZero() {
		return c.deadline, true
	}
	return c.t.Deadline()
}

// Skip delegates request to the parent context. It is a helper method to skip
// the test.
func (c *Context) Skip(args ...any) {
	c.t.Helper()

	c.t.Skip(args...)
}

// Skipf delegates request to the parent context. It is a helper method to skip
// the test with a formatted message.
func (c *Context) Skipf(format string, args ...any) {
	c.t.Helper()

	c.t.Skipf(format, args...)
}

// SkipNow delegates request to the parent context. It is a helper method to skip
// the test immediately.
func (c *Context) SkipNow() {
	c.t.Helper()

	c.t.SkipNow()
}

// Skipped delegates request to the parent context. It reports whether the test
// has been skipped.
func (c *Context) Skipped() bool {
	c.t.Helper()

	return c.t.Skipped()
}

// Log delegates request to the parent context. It provides a logging function
// for the test.
func (c *Context) Log(args ...any) {
	c.t.Helper()

	c.t.Log(args...)
	if c.reporter != nil {
		c.reporter.Log(args...)
	}
}

// Logf delegates request to the parent context. It provides a logging function
// for the test.
func (c *Context) Logf(format string, args ...any) {
	c.t.Helper()

	c.t.Logf(format, args...)
	if c.reporter != nil {
		c.reporter.Logf(format, args...)
	}
}

// Error handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context. Else it delegates the request to the test reporter if available.
func (c *Context) Error(args ...any) {
	c.t.Helper()

	c.failed.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.expect == Success {
		c.t.Error(args...)
	} else if c.reporter != nil {
		c.reporter.Error(args...)
	}
}

// Errorf handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context. Else it delegates the request to the test reporter if available.
func (c *Context) Errorf(format string, args ...any) {
	c.t.Helper()

	c.failed.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.expect == Success {
		c.t.Errorf(format, args...)
	} else if c.reporter != nil {
		c.reporter.Errorf(format, args...)
	}
}

// Fatal handles a fatal failure message that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context. Else it delegates the request to the test
// reporter if available.
func (c *Context) Fatal(args ...any) {
	c.t.Helper()

	c.lockOrExit()
	defer c.unlock()

	if c.expect == Success {
		c.t.Fatal(args...)
	} else if c.reporter != nil {
		c.reporter.Fatal(args...)
	}
	runtime.Goexit()
}

// Fatalf handles a fatal failure message that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context. Else it delegates the request to the test
// reporter if available.
func (c *Context) Fatalf(format string, args ...any) {
	c.t.Helper()

	c.lockOrExit()
	defer c.unlock()

	if c.expect == Success {
		c.t.Fatalf(format, args...)
	} else if c.reporter != nil {
		c.reporter.Fatalf(format, args...)
	}
	runtime.Goexit()
}

// Fail handles a failure message that immediate aborts of the test execution.
// On an expected success, the failure handling is also delegated to the parent
// test context. Else it delegates the request to the test reporter if available.
func (c *Context) Fail() {
	c.t.Helper()

	c.lockOrExit()
	defer c.unlock()

	if c.expect == Success {
		c.t.Fail()
	} else if c.reporter != nil {
		c.reporter.Fail()
	}
	runtime.Goexit()
}

// FailNow handles fatal failure notifications without log output that aborts
// test execution immediately. On an expected success, it the failure handling
// is also delegated to the parent test context. Else it delegates the request
// to the test reporter if available.
func (c *Context) FailNow() {
	c.t.Helper()

	c.lockOrExit()
	defer c.unlock()

	if c.expect == Success {
		c.t.FailNow()
	} else if c.reporter != nil {
		c.reporter.FailNow()
	}
	runtime.Goexit()
}

// Failed reports whether the test has failed.
func (c *Context) Failed() bool {
	c.t.Helper()

	return c.failed.Load()
}

// regexPanic is a regular expression to extract the actual important panic
// stack trace removing the distracting parts from the test framework.
var regexPanic = regexp.MustCompile(`(?m)\nruntime\/debug\.Stack\(\)` +
	`(\n|.)*runtime\/panic\.go:([0-9]+)[^\n]*\n`)

// Panic handles failure notifications of panics that also abort the test
// execution immediately.
func (c *Context) Panic(arg any) {
	c.t.Helper()

	c.lockOrExit()
	defer c.unlock()

	if c.expect == Success {
		stack := regexPanic.Split(string(debug.Stack()), -1)
		c.t.Fatalf("panic: %v\n%s\n%s", arg, stack[0], stack[1])
	} else if c.reporter != nil {
		if reporter, ok := c.reporter.(Panicer); ok {
			reporter.Panic(arg)
		}
	}
	runtime.Goexit()
}

// Run executes the test function in a safe detached environment and checks
// the failure state after the test function has finished. If the expectation
// is not met, a failure is created in the parent test context.
//
// If name is non-empty, a named sub-test is created by delegating to the
// underlying test runner, allowing *Context to be used wherever a named
// sub-test is created using reflect.Run.
func (c *Context) Run(name string, call Func) {
	c.t.Helper()

	if name != "" {
		reflect.Run(c.t, name, call)
		return
	}

	if c.mode == Parallel {
		c.t.Parallel()
	}

	// Register cleanup handlers.
	c.register()

	// Setup shorter deadline for detached test function.
	wait := time.Duration(math.MaxInt64)
	if deadline, ok := c.Deadline(); ok {
		wait = time.Until(deadline)
	}

	// Execute test function with channel to signal completion.
	done := make(chan any, 1)
	go c.run(call, done)

	// Wait for test to finish or deadline to expire.
	select {
	case <-done:
		// Panic is already handled by the reporter.
	case <-time.After(wait):
		c.Fatal("stopped by deadline")
	}
}

// run executes the test function in a safe, detached test environment. The
// function reports execution failure to the parent test context and unlocks
// the waiting test context.
//
// The function is supposed to be called in a goroutine.
func (c *Context) run(test Func, done chan any) {
	c.t.Helper()

	defer func() {
		c.t.Helper()

		// Unlock the waiting test context.
		defer func() { done <- nil }()

		// Intercept and report panic as a failure.
		if arg := recover(); arg != nil {
			c.Panic(arg)
		}
	}()

	test(c)
}

// register registers the clean up handlers with the parent test context.
func (c *Context) register() {
	c.t.Helper()

	// Register cleanup handlers with the parent test context.
	if cu, ok := c.t.(Cleanuper); ok {
		cu.Cleanup(func() {
			c.t.Helper()

			for i := len(c.cleanups) - 1; i >= 0; i-- {
				c.cleanups[i]()
			}
		})
	}

	// Register handler to unlocked the waiting test context.
	c.Cleanup(func() {
		c.t.Helper()
		c.finish()
	})
}

// finish evaluates the final result of the test function in relation to the
// provided expectation.
func (c *Context) finish() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.t.Skipped() {
		return
	}

	switch c.expect {
	case Success:
		if c.failed.Load() {
			c.t.Errorf("Expected test to succeed but it failed: %s", c.Name())
		}
	case Failure:
		if !c.failed.Load() {
			c.t.Errorf("Expected test to fail but it succeeded: %s", c.Name())
		}
	}
}

// lockOrExit either locks the test mutex or aborts a test in case of a pending
// test failure to ensure that only the first failure is reported.
func (c *Context) lockOrExit() {
	c.t.Helper()

	if c.expect == Failure && c.failed.Swap(true) {
		runtime.Goexit()
	}
	c.mu.Lock()
}

// unlock unlocks the wait group of the test by consuming the wait group
// counter completely.
func (c *Context) unlock() {
	if c.wg != nil {
		c.wg.Add(math.MinInt)
	}
	c.mu.Unlock()
}
