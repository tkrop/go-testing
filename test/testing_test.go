package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/sync"
	"github.com/tkrop/go-testing/test"
)

type TestParam struct {
	name     test.Name
	test     func(test.Test)
	expect   test.Expect
	consumed bool
}

var testFailureParams = map[string]TestParam{
	"raw nothing": {
		test:   func(t test.Test) {},
		expect: test.Success,
	},

	"raw errorf": {
		test:   func(t test.Test) { t.Errorf("fail") },
		expect: test.Failure,
	},

	"raw fatalf": {
		test:     func(t test.Test) { t.Fatalf("fail") },
		expect:   test.Failure,
		consumed: true,
	},

	"raw failnow": {
		test:     func(t test.Test) { t.FailNow() },
		expect:   test.Failure,
		consumed: true,
	},

	"raw panic": {
		test:     func(t test.Test) { panic("panic") },
		expect:   test.Failure,
		consumed: true,
	},

	"run success": {
		test: test.InRun(test.Success,
			func(test.Test) {}),
		expect: test.Success,
	},

	"run success with errorf": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.Errorf("fail") }),
		expect: test.Failure,
	},

	"run success with fatalf": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.Fatalf("fail") }),
		expect:   test.Failure,
		consumed: true,
	},

	"run success with failnow": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.FailNow() }),
		expect:   test.Failure,
		consumed: true,
	},

	"run success with panic": {
		test: test.InRun(test.Success,
			func(t test.Test) { panic("panic") }),
		expect:   test.Failure,
		consumed: true,
	},

	"run failure": {
		test: test.InRun(test.Failure,
			func(t test.Test) {}),
		expect: test.Failure,
	},

	"run failure with errorf": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.Errorf("fail") }),
		expect: test.Success,
	},

	"run failure with fatalf": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.Fatalf("fail") }),
		expect:   test.Failure,
		consumed: true,
	},

	"run failure with failnow": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.FailNow() }),
		expect:   test.Failure,
		consumed: true,
	},

	"run failure with panic": {
		test: test.InRun(test.Failure,
			func(t test.Test) { panic("panic") }),
		expect:   test.Failure,
		consumed: true,
	},
}

func testFailures(t test.Test, param TestParam) {
	// Given
	wg := sync.NewLenientWaitGroup()
	t.(*test.Tester).WaitGroup(wg)
	if param.consumed {
		wg.Add(1)
	}

	// When
	param.test(t)

	// Then
	wg.Wait()
	if param.expect == test.Failure {
		assert.Fail(t, "should fail")
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	for name, param := range testFailureParams {
		name, param := name, param
		t.Run(name, test.Run(param.expect, func(t test.Test) {
			testFailures(t, param)
		}))
	}
}

func TestRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range testFailureParams {
		name, param := name, param
		t.Run(name, test.RunSeq(param.expect, func(t test.Test) {
			testFailures(t, param)
		}))
	}
}

func TestNewRun(t *testing.T) {
	test.New[TestParam](t, TestParam{
		test:   func(t test.Test) { t.FailNow() },
		expect: test.Failure,
	}).Run(func(t test.Test, param TestParam) {
		testFailures(t, param)
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	for _, param := range testFailureParams {
		test.New[TestParam](t, TestParam{
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
	}
}

func TestNewNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testFailureParams {
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
	}
}

func TestMap(t *testing.T) {
	test.Map(t, testFailureParams).
		Run(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
}

func TestSlice(t *testing.T) {
	params := make([]TestParam, 0, len(testFailureParams))
	for name, param := range testFailureParams {
		params = append(params, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		})
	}

	test.Slice(t, params).
		Run(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
}

type ParamParam struct {
	name   string
	expect bool
}

func TestNameCastFallback(t *testing.T) {
	test.New[ParamParam](t, ParamParam{name: "value"}).
		Run(func(t test.Test, param ParamParam) {
			assert.Equal(t, t.Name(), "TestNameCastFallback")
		})
}

func TestExpectCastFallback(t *testing.T) {
	test.New[ParamParam](t, ParamParam{expect: false}).
		Run(func(t test.Test, param ParamParam) {})
}

func TestTypePanic(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "not paniced")
		}
	}()
	test.New[TestParam](t, ParamParam{expect: false}).
		Run(func(t test.Test, param TestParam) {})
}
