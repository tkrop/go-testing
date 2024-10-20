package test

import (
	"math"
	"runtime"
	"runtime/debug"
	"strings"
	gosync "sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tkrop/go-testing/internal/slices"
	"github.com/tkrop/go-testing/internal/sync"
)

// Test is a minimal interface for abstracting test methods that are needed to
// setup an isolated test environment for GoMock and Testify.
type Test interface { //nolint:interfacebloat // Minimal interface.
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
	// Log provides a logging function for the test.
	Log(args ...any)
	// Logf provides a logging function for the test.
	Logf(format string, args ...any)
	// Error handles a failure messages when a test is supposed to continue.
	Error(args ...any)
	// Errorf handles a failure messages when a test is supposed to continue.
	Errorf(format string, args ...any)
	// Fatal handles a fatal failure message that immediate aborts of the test
	// execution.
	Fatal(args ...any)
	// Fatalf handles a fatal failure message that immediate aborts of the test
	// execution.
	Fatalf(format string, args ...any)
	// Fail handles a failure message that immediate aborts of the test
	// execution.
	Fail()
	// FailNow handles fatal failure notifications without log output that
	// aborts test execution immediately.
	FailNow()
	// Failed reports whether the test has failed.
	Failed() bool
	// Cleanup is a function called to setup test cleanup after execution.
	Cleanup(cleanup func())
}

// Cleanuper defines an interface to add a custom mehtod that is called after
// the test execution to cleanup the test environment.
type Cleanuper interface {
	Cleanup(cleanup func())
}

// Run creates an isolated (by default) parallel test context running the given
// test function with given expectation. If the expectation is not met, a test
// failure is created in the parent test context.
func Run(expect Expect, test func(Test)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		New(t, expect).Run(test, Parallel)
	}
}

// RunSeq creates an isolated, test context for the given test function with
// given expectation. If the expectation is not met, a test failure is created
// in the parent test context.
func RunSeq(expect Expect, test func(Test)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		New(t, expect).Run(test, !Parallel)
	}
}

// InRun creates an isolated, (by default) sequential test context for the
// given test function with given expectation. If the expectation is not met, a
// test failure is created in the parent test context.
func InRun(expect Expect, test func(Test)) func(Test) {
	return func(t Test) {
		t.Helper()

		New(t, expect).Run(test, !Parallel)
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
}

// New creates a new minimal isolated test context based on the given test
// context with the given expectation. The parent test context is used to
// delegate methods calls to the parent context to propagate test results.
func New(t Test, expect Expect) *Context {
	if tx, ok := t.(*Context); ok {
		return &Context{
			t: tx, wg: tx.wg,
			expect:   expect,
			deadline: tx.deadline,
		}
	}

	return &Context{
		t: t, expect: expect,
		deadline: func(t Test) time.Time {
			defer func() { _ = recover() }()
			deadline, _ := t.Deadline()
			return deadline
		}(t),
	}
}

// Timeout sets up an individual timeout for the test. This does not affect the
// global test timeout or a pending parent timeout that may abort the test, if
// the given duration is exceeding the timeout. A negative or zero duration is
// ignored and will not change the timeout.
func (t *Context) Timeout(timeout time.Duration) *Context {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()

	if timeout > 0 {
		t.deadline = time.Now().Add(timeout)
	}

	return t
}

// StopEarly stops the test by the given duration ahead of the individual or
// global test deadline, to ensure that a cleanup function has sufficient time
// to finish before a global deadline exceeds. The method is not able to extend
// the test deadline. A negative or zero duration is ignored.
//
// Warning: calling this method multiple times will also reduce the deadline
// step by step.
func (t *Context) StopEarly(time time.Duration) *Context {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.deadline.IsZero() && time > 0 {
		t.deadline = t.deadline.Add(-time)
	}

	return t
}

// WaitGroup adds wait group to unlock in case of a failure.
//
//revive:disable-next-line:waitgroup-by-value // own wrapper interface
func (t *Context) WaitGroup(wg sync.WaitGroup) {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.wg = wg
}

// Reporter sets up a test failure reporter. This can be used to validate the
// reported failures in a test environment.
func (t *Context) Reporter(reporter Reporter) {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.reporter = reporter
}

// Cleanup is a function called to setup test cleanup after execution. This
// method is allowing `gomock` to register its `finish` method that reports the
// missing mock calls.
func (t *Context) Cleanup(cleanup func()) {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.cleanups = append(t.cleanups, cleanup)
}

// Name delegates the request to the parent test context.
func (t *Context) Name() string {
	t.t.Helper()

	return t.t.Name()
}

// Helper delegates request to the parent test context.
func (t *Context) Helper() {
	t.t.Helper()
}

// Parallel robustly delegates request to the parent context. It can be called
// multiple times, since it is swallowing the panic that is raised when calling
// `t.Parallel()` multiple times.
func (t *Context) Parallel() {
	t.t.Helper()

	defer func() {
		if err := recover(); err != nil &&
			err != "testing: t.Parallel called multiple times" {
			t.Panic(err)
		}
	}()

	t.t.Parallel()
}

// TempDir delegates the request to the parent test context.
func (t *Context) TempDir() string {
	t.t.Helper()
	return t.t.TempDir()
}

// Setenv delegates request to the parent context, if it is of type
// `*testing.T`. Else it is swallowing the request silently.
func (t *Context) Setenv(key, value string) {
	t.t.Helper()

	t.t.Setenv(key, value)
}

// Deadline delegates request to the parent context. It returns the deadline of
// the test and a flag indicating whether the deadline is set.
func (t *Context) Deadline() (time.Time, bool) {
	t.t.Helper()

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.deadline.IsZero() {
		return t.deadline, true
	}
	return t.t.Deadline()
}

// Skip delegates request to the parent context. It is a helper method to skip
// the test.
func (t *Context) Skip(args ...any) {
	t.t.Helper()

	t.t.Skip(args...)
}

// Skipf delegates request to the parent context. It is a helper method to skip
// the test with a formatted message.
func (t *Context) Skipf(format string, args ...any) {
	t.t.Helper()

	t.t.Skipf(format, args...)
}

// SkipNow delegates request to the parent context. It is a helper method to skip
// the test immediately.
func (t *Context) SkipNow() {
	t.t.Helper()

	t.t.SkipNow()
}

// Skipped delegates request to the parent context. It reports whether the test
// has been skipped.
func (t *Context) Skipped() bool {
	t.t.Helper()

	return t.t.Skipped()
}

// Log delegates request to the parent context. It provides a logging function
// for the test.
func (t *Context) Log(args ...any) {
	t.t.Helper()

	t.t.Log(args...)
}

// Logf delegates request to the parent context. It provides a logging function
// for the test.
func (t *Context) Logf(format string, args ...any) {
	t.t.Helper()

	t.t.Logf(format, args...)
}

// Error handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context. Else it delegates the request to the test reporter if available.
func (t *Context) Error(args ...any) {
	t.t.Helper()

	t.failed.Store(true)

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.expect == Success {
		t.t.Error(args...)
	} else if t.reporter != nil {
		t.reporter.Error(args...)
	}
}

// Errorf handles failure messages where the test is supposed to continue. On
// an expected success, the failure is also delegated to the parent test
// context. Else it delegates the request to the test reporter if available.
func (t *Context) Errorf(format string, args ...any) {
	t.t.Helper()

	t.failed.Store(true)

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.expect == Success {
		t.t.Errorf(format, args...)
	} else if t.reporter != nil {
		t.reporter.Errorf(format, args...)
	}
}

// Fatal handles a fatal failure message that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context. Else it delegates the request to the test
// reporter if available.
func (t *Context) Fatal(args ...any) {
	t.t.Helper()

	if t.failed.Swap(true) {
		runtime.Goexit()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.unlock()

	if t.expect == Success {
		t.t.Fatal(args...)
	} else if t.reporter != nil {
		t.reporter.Fatal(args...)
	}
	runtime.Goexit()
}

// Fatalf handles a fatal failure message that immediate aborts of the test
// execution. On an expected success, the failure handling is also delegated
// to the parent test context. Else it delegates the request to the test
// reporter if available.
func (t *Context) Fatalf(format string, args ...any) {
	t.t.Helper()

	if t.failed.Swap(true) {
		runtime.Goexit()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.unlock()

	if t.expect == Success {
		t.t.Fatalf(format, args...)
	} else if t.reporter != nil {
		t.reporter.Fatalf(format, args...)
	}
	runtime.Goexit()
}

// Fail handles a failure message that immediate aborts of the test execution.
// On an expected success, the failure handling is also delegated to the parent
// test context. Else it delegates the request to the test reporter if available.
func (t *Context) Fail() {
	t.t.Helper()

	if t.failed.Swap(true) {
		runtime.Goexit()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.unlock()

	if t.expect == Success {
		t.t.Fail()
	} else if t.reporter != nil {
		t.reporter.Fail()
	}
	runtime.Goexit()
}

// FailNow handles fatal failure notifications without log output that aborts
// test execution immediately. On an expected success, it the failure handling
// is also delegated to the parent test context. Else it delegates the request
// to the test reporter if available.
func (t *Context) FailNow() {
	t.t.Helper()

	if t.failed.Swap(true) {
		runtime.Goexit()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.unlock()

	if t.expect == Success {
		t.t.FailNow()
	} else if t.reporter != nil {
		t.reporter.FailNow()
	}
	runtime.Goexit()
}

// Failed reports whether the test has failed.
func (t *Context) Failed() bool {
	t.t.Helper()

	return t.failed.Load()
}

// Offset fr original stack in case of panic handling.
//
// TODO: check offset or/and find a better solution to handle panic stack.
const panicOriginStackOffset = 10

// Panic handles failure notifications of panics that also abort the test
// execution immediately.
func (t *Context) Panic(arg any) {
	t.t.Helper()

	if t.failed.Swap(true) {
		runtime.Goexit()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.unlock()

	if t.expect == Success {
		stack := strings.SplitN(string(debug.Stack()),
			"\n", panicOriginStackOffset)
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
func (t *Context) Run(test func(Test), parallel bool) Test {
	t.t.Helper()

	if parallel {
		t.t.Parallel()
	}

	// Register cleanup handlers.
	t.register()

	// Setup shorter deadline for detached test function.
	wait := time.Duration(math.MaxInt64)
	if deadline, ok := t.Deadline(); ok {
		wait = time.Until(deadline)
	}

	// Execute test function with channel to signal completion.
	done := make(chan any, 1)
	go t.run(test, done)

	// Wait for test to finish or deadline to expire.
	select {
	case <-done:
		// Panic is already handled by the reporter.
	case <-time.After(wait):
		t.Fatalf("stopped by deadline")
	}

	return t
}

// run executes the test function in a safe, detached test environment. The
// function reports execution failure to the parent test context and unlocks
// the waiting test context.
//
// The function is supposed to be called in a goroutine.
func (t *Context) run(test func(Test), done chan any) {
	t.t.Helper()

	defer func() {
		t.t.Helper()

		// Unlock the waiting test context.
		defer func() { done <- nil }()

		// Intercept and report panic as a failure.
		if arg := recover(); arg != nil {
			t.Panic(arg)
		}
	}()

	test(t)
}

// register registers the clean up handlers with the parent test context.
func (t *Context) register() {
	t.t.Helper()

	// Register cleanup handlers with the parent test context.
	if c, ok := t.t.(Cleanuper); ok {
		c.Cleanup(func() {
			t.t.Helper()

			t.mu.Lock()
			cleanups := slices.Reverse(t.cleanups)
			t.mu.Unlock()

			for _, cleanup := range cleanups {
				cleanup()
			}
		})
	}

	// Register handler to unlocked the waiting test context.
	t.Cleanup(func() {
		t.t.Helper()
		t.finish()
	})
}

// finish evaluates the final result of the test function in relation to the
// provided expectation.
func (t *Context) finish() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.t.Skipped() {
		return
	}

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

// unlock unlocks the wait group of the test by consuming the wait group
// counter completely.
func (t *Context) unlock() {
	if t.wg != nil {
		t.wg.Add(math.MinInt)
	}
}
