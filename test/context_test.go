package test_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
			ExecTest(t, param)
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
			ExecTest(t, param)
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

// PanicParam is a test parameter type for testing the test context with panic
// cases.
type PanicParam struct {
	parallel bool
	before   func(test.Test)
	during   func(test.Test)
	expect   mock.SetupFunc
}

// testPanicParams is a map of test parameters for testing the test context with
// panic cases.
var testPanicParams = map[string]PanicParam{
	"setenv in run without parallel": {
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
	},

	"setenv in run with parallel": {
		parallel: true,
		during: func(t test.Test) {
			t.Setenv("TESTING", "during")
			assert.Equal(t, "during", os.Getenv("TESTING"))
		},
		expect: test.Panic("testing: test using t.Setenv or t.Chdir" +
			" can not use t.Parallel"),
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
		parallel: true,
		before: func(t test.Test) {
			t.Setenv("TESTING", "before")
			assert.Equal(t, "before", os.Getenv("TESTING"))
		},
		expect: test.Panic("testing: test using t.Setenv or t.Chdir" +
			" can not use t.Parallel"),
	},

	"swallow multiple parallel calls": {
		during: func(t test.Test) {
			t.Parallel()
			t.Parallel()
		},
	},
}

// TestContextPanic is testing the test context with panic cases.
func TestContextPanic(t *testing.T) {
	for name, param := range testPanicParams {
		name, param := name, param
		t.Run(name, test.RunSeq(test.Success, func(t test.Test) {
			// Given
			if param.before != nil {
				mock.NewMocks(t).Expect(param.expect)
				param.before(t)
			}

			// When
			test.New(t, test.Success).Run(func(t test.Test) {
				mock.NewMocks(t).Expect(param.expect)
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
