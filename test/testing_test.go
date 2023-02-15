package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/sync"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type TestParam struct {
	name     test.Name
	setup    mock.SetupFunc
	test     func(test.Test)
	expect   test.Expect
	consumed bool
}

var testParams = map[string]TestParam{
	"base nothing": {
		test:   func(t test.Test) {},
		expect: test.Success,
	},
	"base errorf": {
		test:   func(t test.Test) { t.Errorf("fail") },
		expect: test.Failure,
	},
	"base fatalf": {
		test:     func(t test.Test) { t.Fatalf("fail") },
		expect:   test.Failure,
		consumed: true,
	},
	"base failnow": {
		test:     func(t test.Test) { t.FailNow() },
		expect:   test.Failure,
		consumed: true,
	},
	"base panic": {
		test:     func(t test.Test) { panic("fail") },
		expect:   test.Failure,
		consumed: true,
	},

	"inrun success": {
		test: test.InRun(test.Success,
			func(test.Test) {}),
		expect: test.Success,
	},
	"inrun success with errorf": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.Errorf("fail") }),
		expect: test.Failure,
	},
	"inrun success with fatalf": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.Fatalf("fail") }),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun success with failnow": {
		test: test.InRun(test.Success,
			func(t test.Test) { t.FailNow() }),
		expect:   test.Failure,
		consumed: true,
	},
	"inrun success with panic": {
		test: test.InRun(test.Success,
			func(t test.Test) { panic("fail") }),
		expect:   test.Failure,
		consumed: true,
	},

	"inrun failure": {
		test: test.InRun(test.Failure,
			func(t test.Test) {}),
		expect: test.Failure,
	},
	"inrun failure with errorf": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.Errorf("fail") }),
		expect: test.Success,
	},
	"inrun failure with fatalf": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.Fatalf("fail") }),
		expect:   test.Success,
		consumed: true,
	},
	"inrun failure with failnow": {
		test: test.InRun(test.Failure,
			func(t test.Test) { t.FailNow() }),
		expect:   test.Success,
		consumed: true,
	},
	"inrun failure with panic": {
		test: test.InRun(test.Failure,
			func(t test.Test) { panic("fail") }),
		expect:   test.Success,
		consumed: true,
	},
}

func testFailures(t test.Test, param TestParam) {
	// Given
	if param.setup != nil {
		mock.NewMocks(t).Expect(param.setup)
	}

	wg := sync.NewLenientWaitGroup()
	t.(*test.Tester).WaitGroup(wg)
	if param.consumed {
		wg.Add(1)
	}

	// When
	param.test(t)

	// Then
	wg.Wait()
}

func TestRun(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		name, param := name, param
		t.Run(name, test.Run(param.expect, func(t test.Test) {
			testFailures(t, param)
		}))
	}
}

func TestRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
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

func TestNewRunSeq(t *testing.T) {
	t.Parallel()

	for _, param := range testParams {
		test.New[TestParam](t, TestParam{
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
	}
}

func TestNewRunNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).Run(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
	}
}

func TestNewRunSeqNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
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
	test.Map(t, testParams).
		Run(func(t test.Test, param TestParam) {
			testFailures(t, param)
		})
}

func TestSlice(t *testing.T) {
	params := make([]TestParam, 0, len(testParams))
	for name, param := range testParams {
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
