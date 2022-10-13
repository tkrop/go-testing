package test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const (
	// ExpectSuccess used to express that a test is supposed to succeed.
	ExpectSuccess Expect = true
	// ExpectFailure used to express that a test is supposed to fail.
	ExpectFailure Expect = false
)

// Expect expresses an expection whether a test will succeed or fail.
type Expect bool

// TestingT is  a minimal testing abstraction of `testing.T` that supports
// `testiy` and `gomock`. It can be used as drop in replacement to check
// for expected test failures.
type TestingT struct {
	require.TestingT
	gomock.TestReporter
	t      *testing.T
	expect Expect
	failed bool
}

// NewTestingT creates a new minimal test context based on the given `go-test`
// context.
func NewTestingT(t *testing.T, expect Expect) *TestingT {
	return &TestingT{t: t, expect: expect}
}

// FailNow implements a detached failure handling for `testing.T.FailNow`.
func (m *TestingT) FailNow() {
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.FailNow()
	}
	runtime.Goexit()
}

// Errorf implements a detached failure handling for `testing.T.Errorf`.
func (m *TestingT) Errorf(format string, args ...any) {
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.Errorf(format, args...)
	}
}

// Fatalf implements a detached failure handling for `testing.T.Fatelf`.
func (m *TestingT) Fatalf(format string, args ...any) {
	m.failed = true
	if m.expect == ExpectSuccess {
		m.t.Fatalf(format, args...)
	}
	runtime.Goexit()
}

// test execution the test function in a safe detached environment and check
// the failure state after the test function has finished. If the test result
// is not according to expectation, a failure is created in the parent test
// context.
func (m *TestingT) test(test func(*TestingT)) *TestingT {
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

// Run runs the given test function and checks whether the result is according
// to the expection.
func Run(expect Expect, test func(*TestingT)) func(*testing.T) {
	return func(t *testing.T) {
		NewTestingT(t, expect).test(test)
	}
}

// Failure expects the given test function to fail. If this is the case, the
// failure is intercepted and the test run succeeds.
func Failure(t *testing.T, test func(*TestingT)) {
	NewTestingT(t, ExpectFailure).test(test)
}

// Success expects the given test function to succeed as usually.
func Success(t *testing.T, test func(*TestingT)) {
	NewTestingT(t, ExpectSuccess).test(test)
}
