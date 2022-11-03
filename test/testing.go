package test

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
)

const (
	// ExpectSuccess used to express that a test is supposed to succeed.
	ExpectSuccess Expect = true
	// ExpectFailure used to express that a test is supposed to fail.
	ExpectFailure Expect = false
)

// Expect expresses an expection whether a test will succeed or fail.
type Expect bool

// Test is a minimal interface for abstracting methods on `testing.T` in a
// way that allows to setup an isolated test environment.
type Test interface {
	gomock.TestHelper
	require.TestingT
	Name() string
}

// TestingT is  a minimal testing abstraction of `testing.T` that supports
// `testiy` and `gomock`. It can be used as drop in replacement to check
// for expected test failures.
type TestingT struct {
	Test
	t      Test
	wg     mock.WaitGroup
	expect Expect
	failed bool
}

// NewTestingT creates a new minimal test context based on the given `go-test`
// context.
func NewTestingT(t Test, expect Expect) *TestingT {
	return &TestingT{t: t, expect: expect}
}

// WaitGroup add wait group to unlock in case of a failure.
func (m *TestingT) WaitGroup(wg mock.WaitGroup) {
	m.wg = wg
}

// Name implements a delegate handling for `testing.T.Name`.
func (m *TestingT) Name() string {
	return m.t.Name()
}

// Helper implements a delegate handling for `testing.T.Helper`.
func (m *TestingT) Helper() {
	m.t.Helper()
}

// Unlock unlock wait group of test by consuming the wait group counter
// completely.
func (m *TestingT) Unlock() {
	if m.wg != nil {
		m.wg.Add(math.MinInt)
	}
}

// FailNow implements a detached failure handling for `testing.T.FailNow`.
func (m *TestingT) FailNow() {
	m.t.Helper()
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.FailNow()
	}
	m.Unlock()
	runtime.Goexit()
}

// Errorf implements a detached failure handling for `testing.T.Errorf`.
func (m *TestingT) Errorf(format string, args ...any) {
	m.t.Helper()
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.Errorf(format, args...)
	}
}

// Fatalf implements a detached failure handling for `testing.T.Fatelf`.
func (m *TestingT) Fatalf(format string, args ...any) {
	m.t.Helper()
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.Fatalf(format, args...)
	}
	m.Unlock()
	runtime.Goexit()
}

// test executes the test function in a safe detached environment and check
// the failure state after the test function has finished. If the test result
// is not according to expectation, a failure is created in the parent test
// context.
func (m *TestingT) test(test func(*TestingT)) *TestingT {
	m.t.Helper()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		test(m)
	}()
	wg.Wait()

	switch m.expect {
	case ExpectSuccess:
		require.False(m.t, m.failed,
			fmt.Sprintf("Expected test %s to succeed", m.t.Name()))
	case ExpectFailure:
		require.True(m.t, m.failed,
			fmt.Sprintf("Expected test %s to fail", m.t.Name()))
	}
	return m
}

// Run creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expection.
func Run(expect Expect, test func(*TestingT)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, expect).test(test)
	}
}

// Failure creates an isolaged test environment for the given test function
// and expects the given test function to fail when executed via `t.Run()`. If
// the function fails, the failure is intercepted and the test succeeds.
func Failure(test func(*TestingT)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, ExpectFailure).test(test)
	}
}

// Success creates an isolated test environment for the given test function
// and expects the test to succeed as usually when executed via `t.Run()`. If
// the test failes the result is propagated to the surrounding test.
func Success(test func(*TestingT)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, ExpectSuccess).test(test)
	}
}

// InRun creates an isolated test environment for the given test function with
// given expectation. When executed via `t.Run()` it checks whether the result
// is matching the expection.
func InRun(expect Expect, test func(*TestingT)) func(*TestingT) {
	return func(t *TestingT) {
		NewTestingT(t, expect).test(test)
	}
}

// InFailure creates an isolaged test environment for the given test function
// and expects the test to fail when executed via `t.Run()`. If the test fails,
// the failure is intercepted and the test succeeds.
func InFailure(test func(*TestingT)) func(*TestingT) {
	return func(t *TestingT) {
		NewTestingT(t, ExpectFailure).test(test)
	}
}

// InSuccess creates an isolated test environment for the given test function
// and expects the test to succeed as usually when executed via `t.Run()`. If
// the test failes the result is propagated to the surrounding test.
func InSuccess(test func(*TestingT)) func(*TestingT) {
	return func(t *TestingT) {
		NewTestingT(t, ExpectSuccess).test(test)
	}
}
