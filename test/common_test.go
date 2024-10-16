package test_test

import (
	"regexp"

	"github.com/tkrop/go-testing/internal/sync"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// ParamParam is a test parameter type for the test runner to test evaluation
// of default test parameter names from the test parameter set.
type ParamParam struct {
	name   string
	expect bool
}

// TestParam is a generic test parameter type for testing the test context as
// well as the test runner using the same parameter sets.
type TestParam struct {
	name     test.Name
	setup    mock.SetupFunc
	test     func(test.Test)
	expect   test.Expect
	consumed bool
}

// TestParamMap is a map of test parameters for testing the test context as
// well as the test runner.
type TestParamMap map[string]TestParam

// FilterBy filters the test parameters by the given pattern to test the
// filtering of the test runner.
func (m TestParamMap) FilterBy(pattern string) TestParamMap {
	filter := regexp.MustCompile(pattern)
	params := TestParamMap{}
	for key, value := range m {
		if filter.MatchString(key) {
			params[key] = value
		}
	}
	return params
}

// GetSlice returns the test parameters as a slice of test parameters sets.
func (m TestParamMap) GetSlice() []TestParam {
	params := make([]TestParam, 0, len(m))
	for name, param := range m {
		params = append(params, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		})
	}
	return params
}

var (
	// TestEmpty is a test function that does nothing.
	TestEmpty = func(test.Test) {}
	// TestErrorf is a test function that fails with an error message.
	TestErrorf = func(t test.Test) { t.Errorf("fail") }
	// TestFatalf is a test function that fails with a fatal error message.
	TestFatalf = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.Fatalf("fail") }()
		t.Fatalf("fail")
	}
	// TestFailNow is a test function that fails immediately.
	TestFailNow = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.FailNow() }()
		t.FailNow()
	}
	// TestPanic is a test function that panics.
	TestPanic = func(t test.Test) {
		// Duplicate terminal failures are ignored.
		go func() { t.(*test.Context).Panic("fail") }()
		panic("fail")
	}
)

// testParams is the generic map of test parameters for testing the test
// context as well as the test runner.
var testParams = TestParamMap{
	"base nothing": {
		test:   TestEmpty,
		expect: test.Success,
	},
	"base errorf": {
		test:   TestErrorf,
		expect: test.Failure,
	},
	"base fatalf": {
		test:     TestFatalf,
		expect:   test.Failure,
		consumed: true,
	},
	"base failnow": {
		test:     TestFailNow,
		expect:   test.Failure,
		consumed: true,
	},
	"base panic": {
		test:     TestPanic,
		expect:   test.Failure,
		consumed: true,
	},

	"inrun success": {
		test:   test.InRun(test.Success, TestEmpty),
		expect: test.Success,
	},
	"inrun success with errorf": {
		test:   test.InRun(test.Success, TestErrorf),
		expect: test.Failure,
	},
	"inrun success with fatalf": {
		test:     test.InRun(test.Success, TestFatalf),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun success with failnow": {
		test:     test.InRun(test.Success, TestFailNow),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun success with panic": {
		test:     test.InRun(test.Success, TestPanic),
		expect:   test.Failure,
		consumed: true,
	},

	"inrun failure": {
		test:   test.InRun(test.Failure, TestEmpty),
		expect: test.Failure,
	},
	"inrun failure with errorf": {
		test:   test.InRun(test.Failure, TestErrorf),
		expect: test.Success,
	},
	"inrun failure with fatalf": {
		test:     test.InRun(test.Failure, TestFatalf),
		expect:   test.Success,
		consumed: true,
	},
	"inrun failure with failnow": {
		test:     test.InRun(test.Failure, TestFailNow),
		expect:   test.Success,
		consumed: true,
	},
	"inrun failure with panic": {
		test:     test.InRun(test.Failure, TestPanic),
		expect:   test.Success,
		consumed: true,
	},
}

// ExecTest is the generic function to execute a test with the given test
// parameters.
func ExecTest(t test.Test, param TestParam) {
	// Given
	if param.setup != nil {
		mock.NewMocks(t).Expect(param.setup)
	}

	wg := sync.NewLenientWaitGroup()
	t.(*test.Context).WaitGroup(wg)
	if param.consumed {
		wg.Add(1)
	}

	// When
	param.test(t)

	// Then
	wg.Wait()
}
