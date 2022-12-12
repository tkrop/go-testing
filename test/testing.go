package test

import (
	"math"
	"runtime"
	gosync "sync"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/sync"
	"github.com/tkrop/testing/utils/slices"
)

const (
	// ExpectSuccess used to express that a test is supposed to succeed.
	ExpectSuccess Expect = true
	// ExpectFailure used to express that a test is supposed to fail.
	ExpectFailure Expect = false
)

// Expect expresses an expectation whether a test will succeed or fail.
type Expect bool

// Test is a minimal interface for abstracting methods on `testing.T` in a
// way that allows to setup an isolated test environment.
type Test interface {
	gomock.TestHelper
	require.TestingT
	Parallel()
	Name() string
}

// Cleanuper defines an interface to add a custom mehtod that is called after
// the test execution to cleanup the test environment.
type Cleanuper interface {
	Cleanup(f func())
}

// TestingT is  a minimal testing abstraction of `testing.T` that supports
// `testiy` and `gomock`. It can be used as drop in replacement to check for
// expected test failures.
type TestingT struct {
	Test
	sync.Synchronizer
	t        Test
	wg       sync.WaitGroup
	mu       gosync.Mutex
	failed   atomic.Bool
	cleanups []func()
	expect   Expect
}

// NewTestingT creates a new minimal test context based on the given `go-test`
// context.
func NewTestingT(t Test, expect Expect) *TestingT {
	if tx, ok := t.(*TestingT); ok {
		return &TestingT{t: tx, wg: tx.wg, expect: expect}
	}
	return &TestingT{t: t, expect: expect}
}

// Parallel delegates request to `testing.T.Parallel()`.
func (m *TestingT) Parallel() {
	m.t.Parallel()
}

// Cleanup is a function called to setup test cleanup after execution. This
// method is allowing `gomock` to register its `finish` method that reports the
// missing mock calls.
func (m *TestingT) Cleanup(cleanup func()) {
	m.mu.Lock()
	m.cleanups = append(m.cleanups, cleanup)
	m.mu.Unlock()
}

// WaitGroup adds wait group to unlock in case of a failure.
func (m *TestingT) WaitGroup(wg sync.WaitGroup) {
	m.wg = wg
}

// Name provides the test name by delegating the request to `testing.T.Name()`.
func (m *TestingT) Name() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.t.Name()
}

// Helper delegates request to `testing.T.Helper()`.
func (m *TestingT) Helper() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t.Helper()
}

// Unlock unlocks the wait group of the test by consuming the wait group
// counter completely.
func (m *TestingT) Unlock() {
	if m.wg != nil {
		m.wg.Add(math.MinInt)
	}
}

// FailNow provides the detached failure handling for `testing.T.FailNow()`.
func (m *TestingT) FailNow() {
	m.Helper()
	m.failed.Store(true)
	defer m.Unlock()
	if m.expect == ExpectSuccess {
		m.t.FailNow()
	}
	runtime.Goexit()
}

// Errorf provides the detached failure handling for `testing.T.Errorf()`.
func (m *TestingT) Errorf(format string, args ...any) {
	m.Helper()
	m.failed.Store(true)
	if m.expect == ExpectSuccess {
		m.t.Errorf(format, args...)
	}
}

// Fatalf provides the detached failure handling for `testing.T.Fatelf()`.
func (m *TestingT) Fatalf(format string, args ...any) {
	m.Helper()
	m.failed.Store(true)
	defer m.Unlock()
	if m.expect == ExpectSuccess {
		m.t.Fatalf(format, args...)
	}
	runtime.Goexit()
}

// Run executes the test function in a safe detached environment and check
// the failure state after the test function has finished. If the test result
// is not according to expectation, a failure is created in the parent test
// context.
func (m *TestingT) Run(test func(Test)) Test {
	m.Helper()

	// register cleanup handlers.
	m.Cleanup(func() {
		m.Helper()
		m.finish()
	})
	if c, ok := m.t.(Cleanuper); ok {
		c.Cleanup(func() {
			m.Helper()
			m.cleanup()
		})
	}

	// execute test function.
	wg := sync.NewWaitGroup()
	wg.Add(1)
	go func() {
		defer wg.Done()
		test(m)
	}()
	wg.Wait()

	return m
}

// cleanup runs the cleanup methods registered on the isolated test environment.
func (m *TestingT) cleanup() {
	m.mu.Lock()
	cleanups := slices.Reverse(m.cleanups)
	m.mu.Unlock()

	for _, cleanup := range cleanups {
		cleanup()
	}
}

// finish evaluates the final result of the test function in relation to the
// provided expectation.
func (m *TestingT) finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch m.expect {
	case ExpectSuccess:
		if m.failed.Load() {
			m.t.Errorf("Expected test to succeed but it failed: %s", m.t.Name())
		}
	case ExpectFailure:
		if !m.failed.Load() {
			m.t.Errorf("Expected test to fail but it succeeded: %s", m.t.Name())
		}
	}
}

// Run creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expectation.
func Run(expect Expect, test func(Test)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, expect).Run(test)
	}
}

// Failure creates an isolaged test environment for the given test function
// and expects the given test function to fail when executed via `t.Run()`. If
// the function fails, the failure is intercepted and the test succeeds.
func Failure(test func(Test)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, ExpectFailure).Run(test)
	}
}

// Success creates an isolated test environment for the given test function
// and expects the test to succeed as usually when executed via `t.Run()`. If
// the test failes the result is propagated to the surrounding test.
func Success(test func(Test)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, ExpectSuccess).Run(test)
	}
}

// InRun creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expectation.
func InRun(expect Expect, test func(Test)) func(Test) {
	return func(t Test) {
		NewTestingT(t, expect).Run(test)
	}
}

// InFailure creates an isolaged test environment for the given test function
// and expects the test to fail when executed via `t.Run()`. If the test fails,
// the failure is intercepted and the test succeeds.
func InFailure(test func(Test)) func(Test) {
	return func(t Test) {
		NewTestingT(t, ExpectFailure).Run(test)
	}
}

// InSuccess creates an isolated test environment for the given test function
// and expects the test to succeed as usually when executed via `t.Run()`. If
// the test failes the result is propagated to the surrounding test.
func InSuccess(test func(Test)) func(Test) {
	return func(t Test) {
		NewTestingT(t, ExpectSuccess).Run(test)
	}
}
