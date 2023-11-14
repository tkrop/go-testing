package test_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	finished := false
	test.New[TestParam](t, TestParam{
		test:   func(t test.Test) { t.FailNow() },
		expect: test.Failure,
	}).Run(func(t test.Test, param TestParam) {
		defer func() { finished = true }()
		testFailures(t, param)
	}).Cleanup(func() {
		require.True(t, finished)
	})
}

func TestNewRunSeq(t *testing.T) {
	t.Parallel()

	for _, param := range testParams {
		finished := false
		test.New[TestParam](t, TestParam{
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			testFailures(t, param)
		}).Cleanup(func() {
			require.True(t, finished)
		})
	}
}

func TestNewRunNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).Run(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			testFailures(t, param)
		}).Cleanup(func() {
			require.True(t, finished)
		})
	}
}

func TestNewRunSeqNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			testFailures(t, param)
		}).Cleanup(func() {
			require.True(t, finished)
		})
	}
}

func TestMap(t *testing.T) {
	count := atomic.Int32{}

	test.Map(t, testParams).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			testFailures(t, param)
		}).
		Cleanup(func() {
			require.Equal(t, len(testParams), int(count.Load()))
		})
}

func TestSlice(t *testing.T) {
	count := atomic.Int32{}

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
			defer count.Add(1)
			testFailures(t, param)
		}).
		Cleanup(func() {
			require.Equal(t, len(testParams), int(count.Load()))
		})
}

type ParamParam struct {
	name   string
	expect bool
}

func TestTempDir(t *testing.T) {
	test.New[ParamParam](t, ParamParam{expect: true}).
		Run(func(t test.Test, param ParamParam) {
			assert.NotEmpty(t, t.TempDir())
		})
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
