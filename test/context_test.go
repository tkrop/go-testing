package test_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/internal/sync"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// TestRun is testing the test context with single test cases running in
// parallel.
func TestRun(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		name, param := name, param
		t.Run(name, test.Run(param.expect, func(t test.Test) {
			param.CheckName(t)
			param.ExecTest(t)
		}))
	}
}

// TestRunSeq is testing the test context with single test cases running in
// sequence.
func TestRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		name, param := name, param
		t.Run(name, test.RunSeq(param.expect, func(t test.Test) {
			param.CheckName(t)
			param.ExecTest(t)
		}))
	}
}

// TestTempDir is testing the test context creating temporary directory.
func TestTempDir(t *testing.T) {
	t.Parallel()

	t.Run("create", test.Run(test.Success, func(t test.Test) {
		assert.NotEmpty(t, t.TempDir())
	}))
}

// ContextParam is a test parameter type for testing the test context.
type ContextParam struct {
	setup mock.SetupFunc
	test  test.Func
}

// testContextParams is a map of test parameters for testing the test context.
var testContextParams = map[string]ContextParam{
	"panic": {
		setup: mock.Chain(
			test.Fatalf("panic: %v\n%s\n%s", "test", gomock.Any(), gomock.Any()),
		),
		test: func(test.Test) {
			panic("test")
		},
	},
	// TODO: add more test cases for the test context.
}

// TestContext is testing the test context with single simple test cases.
func TestContext(t *testing.T) {
	for name, param := range testContextParams {
		name, param := name, param
		t.Run(name, test.Run(test.Success, func(t test.Test) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			test.New(t, test.Success).
				Run(param.test, !test.Parallel)
		}))
	}
}

// CleanupParam is a test parameter type for testing the Cleanup method.
type CleanupParam struct {
	test test.Func
	wait int
}

// testCleanupParams is a map of test parameters for testing the Cleanup method.
var testCleanupParams = map[string]CleanupParam{
	"nil cleanup": {
		test: func(t test.Test) {
			t.Cleanup(nil)
		},
	},
	"single cleanup": {
		test: func(t test.Test) {
			t.Cleanup(func() { t.(*test.Context).Done() })
		},
		wait: 1,
	},
	"multiple cleanups": {
		test: func(t test.Test) {
			t.Cleanup(func() { t.(*test.Context).Done() })
			t.Cleanup(func() { t.(*test.Context).Done() })
			t.Cleanup(func() { t.(*test.Context).Done() })
		},
		wait: 3,
	},
	"cleanup with nil mixed": {
		test: func(t test.Test) {
			t.Cleanup(nil)
			t.Cleanup(func() { t.(*test.Context).Done() })
			t.Cleanup(nil)
		},
		wait: 1,
	},
}

// TestCleanup is testing the Cleanup method with various scenarios including nil input.
func TestCleanup(t *testing.T) {
	for name, param := range testCleanupParams {
		name, param := name, param
		t.Run(name, test.Run(test.Success, func(t test.Test) {
			// Given
			wg := sync.NewWaitGroup()
			wg.Add(param.wait + 1)
			t.(*test.Context).WaitGroup(wg)
			t.Cleanup(func() { wg.Wait() })

			// When
			test.New(t, test.Success).
				Run(param.test, test.Parallel)

			// Then
			defer wg.Done()
		}))
	}
}

// ParallelParam is a test parameter type for testing the test context in
// conflicting parallel cases resulting in panics.
type ParallelParam struct {
	setup    mock.SetupFunc
	parallel bool
	before   test.SetupFunc
	during   test.Func
}

// testParallelParams is a map of test parameters for testing the test context
// in conflicting parallel cases resulting in a panics.
var testParallelParams = map[string]ParallelParam{
	"setenv in run without parallel": {
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv in run with parallel": {
		setup: test.Panic("testing: test using t.Setenv or t.Chdir" +
			" can not use t.Parallel"),
		parallel: true,
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv before run without parallel": {
		before: func(t test.Test) {
			t.Setenv("TESTING", "before")
			assert.Equal(t, "before", os.Getenv("TESTING"))
		},
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv before run with parallel": {
		setup: test.Panic("testing: test using t.Setenv or t.Chdir" +
			" can not use t.Parallel"),
		parallel: true,
		before: func(t test.Test) {
			t.Setenv("TESTING", "before")
			assert.Equal(t, "before", os.Getenv("TESTING"))
		},
	},

	"swallow multiple parallel calls": {
		during: func(t test.Test) {
			t.Parallel()
			t.Parallel()
		},
	},
}

// TestContextParallel is testing the test context in conflicting parallel
// cases creating panics.
func TestContextParallel(t *testing.T) {
	for name, param := range testParallelParams {
		name, param := name, param
		t.Run(name, test.RunSeq(test.Success, func(t test.Test) {
			// Given
			if param.before != nil {
				mock.NewMocks(t).Expect(param.setup)
				param.before(t)
			}

			// When
			test.New(t, test.Success).Run(func(t test.Test) {
				mock.NewMocks(t).Expect(param.setup)
				param.during(t)
			}, param.parallel)
		}))
	}
}

type TestDeadlineParam struct {
	time, early, sleep time.Duration
	expect             mock.SetupFunc
	failure            test.Expect
}

var TestDeadlineParams = map[string]TestDeadlineParam{
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

func TestDeadline(t *testing.T) {
	t.Parallel()

	test.Map(t, TestDeadlineParams).
		Timeout(0).StopEarly(0).
		Run(func(t test.Test, param TestDeadlineParam) {
			mock.NewMocks(t).Expect(param.expect)

			test.New(t, !param.failure).
				Timeout(param.time).StopEarly(param.early).
				Run(func(t test.Test) {
					// When
					time.Sleep(param.sleep)

					// Then
					t.Fatalf("finished regularly")
				}, !test.Parallel)
		})
}
