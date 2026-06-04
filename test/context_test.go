package test_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/internal/sync"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/reflect"
	"github.com/tkrop/go-testing/test"
)

// paramParams is a test parameter type for the test runner to test evaluation
// of default test parameter names from the test parameter set.
type paramParams struct {
	name   string
	expect bool
}

// CheckName checks if the test name contains the expected name. It is used to
// verify that the test name is correctly set in the test runner.
func (p *paramParams) CheckName(t test.Test) {
	assert.Contains(t, t.Name(),
		strings.ReplaceAll(p.name, " ", "-"))
}

// testParams is a generic test parameter type for testing the test context as
// well as the test runner using the same parameter sets.
type testParams struct {
	name     string
	setup    mock.SetupFunc
	test     test.Func
	expect   test.Expect
	consumed bool
}

// Rename returns a new test parameter set with the given name.
func (p *testParams) Rename(name string) testParams {
	return testParams{
		name:     name,
		setup:    p.setup,
		test:     p.test,
		expect:   p.expect,
		consumed: p.consumed,
	}
}

// Rename returns a new test parameter set with the given name.
func (p *testParams) Copy() testParams {
	return testParams{
		name:     p.name,
		setup:    p.setup,
		test:     p.test,
		expect:   p.expect,
		consumed: p.consumed,
	}
}

// CheckName checks if the test name contains the expected name.
// It is used to verify that the test name is correctly set in the test runner.
func (p *testParams) CheckName(t test.Test) {
	assert.Contains(t, t.Name(),
		strings.ReplaceAll(p.name, " ", "-"))
}

// ExecTest is the generic function to execute a test with the given test
// parameters.
func (p *testParams) ExecTest(t test.Test) {
	// Given
	if p.setup != nil {
		mock.NewMocks(t).Expect(p.setup)
	}

	wg := sync.NewLenientWaitGroup()
	test.Cast[*test.Context](t).WaitGroup(wg)
	if p.consumed {
		wg.Add(1)
	}

	// When
	p.test(t)

	// Then
	wg.Wait()
	if p.expect == test.Failure {
		assert.True(t, t.Failed())
	}
}

// testParamMap is a map of test parameters for testing the test context as
// well as the test runner.
type testParamMap map[string]testParams

// FilterBy filters the test parameters by the given pattern to test the
// filtering of the test runner.
func (m testParamMap) FilterBy(pattern string) testParamMap {
	filter := regexp.MustCompile(pattern)
	params := testParamMap{}
	for key, value := range m {
		if filter.MatchString(key) {
			params[key] = value
		}
	}
	return params
}

// GetSlice returns the test parameters as a slice of test parameters sets.
func (m testParamMap) GetSlice() []testParams {
	params := make([]testParams, 0, len(m))
	for name, param := range m {
		params = append(params, testParams{
			name:   name,
			test:   param.test,
			expect: param.expect,
		})
	}
	return params
}

var (
	// TestEmpty is a test function that does nothing.
	TestEmpty = func(test.Test) {}
	// TestSkip is a test function that skips the test.
	TestSkip = func(t test.Test) { t.Skip("skip") }
	// TestSkipf is a test function that skips the test with a formatted message.
	TestSkipf = func(t test.Test) { t.Skipf("%s", "skip") }
	// TestSkipNow is a test function that skips the test immediately.
	TestSkipNow = func(t test.Test) { t.SkipNow() }
	// TestLog is a test function that logs a message.
	TestLog = func(t test.Test) { t.Log("log") }
	// TestLogf is a test function that logs a formatted message.
	TestLogf = func(t test.Test) { t.Logf("%s", "log") }
	// TestError is a test function that fails with an error message.
	TestError = func(t test.Test) { t.Error("fail") }
	// TestErrorf is a test function that fails with a formatted error message.
	TestErrorf = func(t test.Test) { t.Errorf("%s", "fail") }
	// TestFatal is a test function that fails with a fatal error message.
	TestFatal = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.Fatal("fail") }()
		t.Fatal("fail")
	}
	// TestFatalf is a test function that fails with a fatal formatted error
	// message.
	TestFatalf = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.Fatalf("%s", "fail") }()
		t.Fatalf("%s", "fail")
	}
	// TestFail is a test function that fails.
	TestFail = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.Fail() }()
		t.Fail()
	}
	// TestFailNow is a test function that fails immediately.
	TestFailNow = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.FailNow() }()
		t.FailNow()
	}
	// TestPanic is a test function that panics.
	TestPanic = func(test.Test) { panic("fail") }
	// CleanupEmpty is a cleanup function that does nothing.
	CleanupEmpty = func() {}
	// CleanupPanic is a cleanup function that panics.
	CleanupPanic = func() { panic("cleanup") }
)

// commonTestCases is the generic map of test parameters for testing the test
// context as well as the test runner.
var commonTestCases = testParamMap{
	"": {},
	"base-nothing": {
		test:   TestEmpty,
		expect: test.Success,
	},
	"base-skip": {
		test:   TestSkip,
		expect: test.Success,
	},
	"base-skipf": {
		test:   TestSkipf,
		expect: test.Success,
	},
	"base-skipnow": {
		test:   TestSkipNow,
		expect: test.Success,
	},
	"base-log": {
		test:   TestLog,
		expect: test.Success,
	},
	"base-logf": {
		test:   TestLogf,
		expect: test.Success,
	},
	"base-error": {
		test:   TestError,
		expect: test.Failure,
	},
	"base-errorf": {
		test:   TestErrorf,
		expect: test.Failure,
	},
	"base-fatal": {
		test:     TestFatal,
		expect:   test.Failure,
		consumed: true,
	},
	"base-fatalf": {
		test:     TestFatalf,
		expect:   test.Failure,
		consumed: true,
	},
	"base-fail": {
		test:     TestFail,
		expect:   test.Failure,
		consumed: true,
	},
	"base-failnow": {
		test:     TestFailNow,
		expect:   test.Failure,
		consumed: true,
	},
	"base-panic": {
		test:     TestPanic,
		expect:   test.Failure,
		consumed: true,
	},

	"inrun-success": {
		test:   test.Wrap(test.Success, TestEmpty),
		expect: test.Success,
	},
	"inrun-success-with-skip": {
		test:   test.Wrap(test.Success, TestSkip),
		expect: test.Success,
	},
	"inrun-success-with-skipf": {
		test:   test.Wrap(test.Success, TestSkipf),
		expect: test.Success,
	},
	"inrun-success-with-skipnow": {
		test:   test.Wrap(test.Success, TestSkipNow),
		expect: test.Success,
	},
	"inrun-success-with-log": {
		test:   test.Wrap(test.Success, TestLog),
		expect: test.Success,
	},
	"inrun-success-with-logf": {
		test:   test.Wrap(test.Success, TestLogf),
		expect: test.Success,
	},
	"inrun-success-with-error": {
		test:   test.Wrap(test.Success, TestError),
		expect: test.Failure,
	},
	"inrun-success-with-errorf": {
		test:   test.Wrap(test.Success, TestErrorf),
		expect: test.Failure,
	},
	"inrun-success-with-fatal": {
		test:     test.Wrap(test.Success, TestFatal),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun-success-with-fatalf": {
		test:     test.Wrap(test.Success, TestFatalf),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun-success-with-fail": {
		test:     test.Wrap(test.Success, TestFail),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun-success-with-failnow": {
		test:     test.Wrap(test.Success, TestFailNow),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun-success-with-panic": {
		test:     test.Wrap(test.Success, TestPanic),
		expect:   test.Failure,
		consumed: true,
	},

	"inrun-failure": {
		test:   test.Wrap(test.Failure, TestEmpty),
		expect: test.Failure,
	},
	"inrun-failure-with-skip": {
		test:   test.Wrap(test.Failure, TestSkip),
		expect: test.Failure,
	},
	"inrun-failure-with-skipf": {
		test:   test.Wrap(test.Failure, TestSkipf),
		expect: test.Failure,
	},
	"inrun-failure-with-skipnow": {
		test:   test.Wrap(test.Failure, TestSkipNow),
		expect: test.Failure,
	},
	"inrun-failure-with-log": {
		test:   test.Wrap(test.Failure, TestLog),
		expect: test.Failure,
	},
	"inrun-failure-with-logf": {
		test:   test.Wrap(test.Failure, TestLogf),
		expect: test.Failure,
	},
	"inrun-failure-with-error": {
		test:   test.Wrap(test.Failure, TestError),
		expect: test.Success,
	},
	"inrun-failure-with-errorf": {
		test:   test.Wrap(test.Failure, TestErrorf),
		expect: test.Success,
	},
	"inrun-failure-with-fatal": {
		test:     test.Wrap(test.Failure, TestFatal),
		expect:   test.Success,
		consumed: true,
	},
	"inrun-failure-with-fatalf": {
		test:     test.Wrap(test.Failure, TestFatalf),
		expect:   test.Success,
		consumed: true,
	},
	"inrun-failure-with-fail": {
		test:     test.Wrap(test.Failure, TestFail),
		expect:   test.Success,
		consumed: true,
	},
	"inrun-failure-with-failnow": {
		test:     test.Wrap(test.Failure, TestFailNow),
		expect:   test.Success,
		consumed: true,
	},
	"inrun-failure-with-panic": {
		test:     test.Wrap(test.Failure, TestPanic),
		expect:   test.Success,
		consumed: true,
	},
}

// TestRun is testing the test context with single test cases running in
// parallel.
func TestRun(t *testing.T) {
	t.Parallel()

	for name, param := range commonTestCases {
		t.Run(name, test.Run(func(t test.Tester) {
			// Given
			t.Expect(param.expect)

			// Then
			param.CheckName(t)
			param.ExecTest(t)
		}))
	}
}

// TestRunSeq is testing the test context with single test cases running in
// sequence - but still running in parallel.
func TestRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range commonTestCases {
		t.Run(name, test.RunSeq(func(t test.Tester) {
			t.Parallel()

			// Given
			t.Expect(param.expect)

			// Then
			param.CheckName(t)
			param.ExecTest(t)
		}))
	}
}

// TestTempDir is testing the test context creating temporary directory.
func TestTempDir(t *testing.T) {
	t.Parallel()

	t.Run("create", test.Run(func(t test.Tester) {
		// Then
		assert.NotEmpty(t, t.TempDir())
	}))
}

// contextParams is a test parameter type for testing the test context.
type contextParams struct {
	setup mock.SetupFunc
	test  test.Func
}

// contextTestCases is a map of test parameters for testing the test context.
var contextTestCases = map[string]contextParams{
	"panic": {
		setup: mock.Chain(
			test.Fatalf("panic: %v\n%s\n%s", "test", gomock.Any(), gomock.Any()),
		),
		test: func(test.Test) {
			panic("test")
		},
	},

	// Sub-context copy-semantics cases.
	"success-sequential": {
		test: func(t test.Test) {
			parent := test.New(t).Mode(test.Sequential).Expect(test.Success)
			child := test.New(parent)
			getter := reflect.NewGetter(child)

			assert.Equal(t, test.Success, getter.Get("expect"))
			assert.Equal(t, test.Sequential, getter.Get("mode"))
		},
	},
	"success-parallel": {
		test: func(t test.Test) {
			parent := test.New(t).Mode(test.Parallel).Expect(test.Success)
			child := test.New(parent)
			getter := reflect.NewGetter(child)

			assert.Equal(t, test.Success, getter.Get("expect"))
			assert.Equal(t, test.Parallel, getter.Get("mode"))
		},
	},
	"failure-sequential": {
		test: func(t test.Test) {
			parent := test.New(t).Mode(test.Sequential).Expect(test.Failure)
			child := test.New(parent)
			getter := reflect.NewGetter(child)

			assert.Equal(t, test.Failure, getter.Get("expect"))
			assert.Equal(t, test.Sequential, getter.Get("mode"))
		},
	},
	"failure-parallel": {
		test: func(t test.Test) {
			parent := test.New(t).Mode(test.Parallel).Expect(test.Failure)
			child := test.New(parent)
			getter := reflect.NewGetter(child)

			assert.Equal(t, test.Failure, getter.Get("expect"))
			assert.Equal(t, test.Parallel, getter.Get("mode"))
		},
	},
}

// TestContext is testing the test context with single simple test cases.
func TestContext(t *testing.T) {
	t.Parallel()

	for name, param := range contextTestCases {
		t.Run(name, test.Run(func(t test.Tester) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			test.New(t).Mode(test.Sequential).
				Expect(test.Success).Run("", param.test)
		}))
	}
}

// cleanupParams is a test parameter type for testing the Cleanup method.
type cleanupParams struct {
	test test.Func
	wait int
}

// cleanupTestCases is a map of test parameters for testing the Cleanup method.
var cleanupTestCases = map[string]cleanupParams{
	"nil-cleanup": {
		test: func(t test.Test) {
			t.Cleanup(nil)
		},
	},
	"single-cleanup": {
		test: func(t test.Test) {
			t.Cleanup(func() { test.Cast[*test.Context](t).Done() })
		},
		wait: 1,
	},
	"multiple-cleanups": {
		test: func(t test.Test) {
			t.Cleanup(func() { test.Cast[*test.Context](t).Done() })
			t.Cleanup(func() { test.Cast[*test.Context](t).Done() })
			t.Cleanup(func() { test.Cast[*test.Context](t).Done() })
		},
		wait: 3,
	},
	"cleanup-with-nil-mixed": {
		test: func(t test.Test) {
			t.Cleanup(nil)
			t.Cleanup(func() { test.Cast[*test.Context](t).Done() })
			t.Cleanup(nil)
		},
		wait: 1,
	},
}

// TestCleanup is testing the Cleanup method with various scenarios including nil input.
func TestCleanup(t *testing.T) {
	t.Parallel()

	for name, param := range cleanupTestCases {
		t.Run(name, test.Run(func(t test.Tester) {
			// Given
			wg := sync.NewWaitGroup()
			wg.Add(param.wait + 1)
			test.Cast[*test.Context](t).WaitGroup(wg)
			t.Cleanup(func() { wg.Wait() })

			// When
			test.New(t).Mode(test.Parallel).
				Expect(test.Success).Run("", param.test)

			// Then
			defer wg.Done()
		}))
	}
}

// parallelParams is a test parameter type for testing the test context in
// conflicting parallel cases resulting in panics.
type parallelParams struct {
	setup  mock.SetupFunc
	mode   test.Mode
	before test.SetupFunc
	during test.Func
}

// parallelTestCases is a map of test parameters for testing the test context
// in conflicting parallel cases resulting in a panics.
var parallelTestCases = map[string]parallelParams{
	"setenv-in-run-without-parallel": {
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv-in-run-with-parallel": {
		setup: test.Panic("testing: test using t.Setenv, t.Chdir, or " +
			"cryptotest.SetGlobalRandom can not use t.Parallel"),
		mode: test.Parallel,
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv-before-run-without-parallel": {
		before: func(t test.Test) {
			t.Setenv("TESTING", "before")
			assert.Equal(t, "before", os.Getenv("TESTING"))
		},
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv-before-run-with-parallel": {
		setup: test.Panic("testing: test using t.Setenv, t.Chdir, or " +
			"cryptotest.SetGlobalRandom can not use t.Parallel"),
		mode: test.Parallel,
		before: func(t test.Test) {
			t.Setenv("TESTING", "before")
			assert.Equal(t, "before", os.Getenv("TESTING"))
		},
	},

	"swallow-multiple-parallel-calls": {
		during: func(t test.Test) {
			t.Parallel()
			t.Parallel()
		},
	},
}

// TestContextParallel is testing the test context in conflicting parallel
// cases creating panics.
func TestContextParallel(t *testing.T) {
	for name, param := range parallelTestCases {
		t.Run(name, test.RunSeq(func(t test.Tester) {
			// Given
			if param.before != nil {
				mock.NewMocks(t).Expect(param.setup)
				param.before(t)
			}

			// When
			test.New(t).Mode(param.mode).
				Expect(test.Success).Run("", func(t test.Test) {
				mock.NewMocks(t).Expect(param.setup)
				param.during(t)
			})
		}))
	}
}

// deadlineParams is a test parameter type for testing the Deadline method.
type deadlineParams struct {
	time, early, sleep time.Duration
	expect             mock.SetupFunc
	failure            test.Expect
}

// deadlineTestCases is a map of test parameters for testing the Deadline
// method.
var deadlineTestCases = map[string]deadlineParams{
	"failed": {
		time:    0,
		early:   0,
		sleep:   time.Millisecond,
		expect:  test.Fatalf("finished regularly"),
		failure: test.Failure,
	},
	"timeout": {
		time:    time.Millisecond,
		early:   0,
		sleep:   time.Second,
		expect:  test.Fatalf("stopped by deadline"),
		failure: test.Failure,
	},
	"early": {
		time:    5 * time.Millisecond,
		early:   4 * time.Millisecond,
		sleep:   4 * time.Millisecond,
		expect:  test.Fatalf("stopped by deadline"),
		failure: test.Failure,
	},
	"to-late": {
		time:    5 * time.Millisecond,
		early:   1 * time.Millisecond,
		sleep:   1 * time.Millisecond,
		expect:  test.Fatalf("finished regularly"),
		failure: test.Failure,
	},
	"parent": {
		time:  0,
		early: 0,
		sleep: 12 * time.Millisecond,
		expect: mock.Chain(
			test.Fatalf("stopped by deadline"),
			test.Errorf("Expected test to succeed but it failed: %s",
				"TestDeadline/parent"),
		),
		failure: test.Failure,
	},
}

// TestDeadline is testing the Deadline method with various scenarios including
// timeouts and early stops.
func TestDeadline(t *testing.T) {
	test.Map(t, deadlineTestCases).
		Timeout(0).StopEarly(0).
		Run(func(t test.Test, param deadlineParams) {
			mock.NewMocks(t).Expect(param.expect)

			test.New(t).Mode(test.Sequential).Expect(!param.failure).
				Timeout(param.time).StopEarly(param.early).
				Run("", func(t test.Test) {
					// When
					time.Sleep(param.sleep)

					// Then
					t.Fatal("finished regularly")
				})
		})
}
