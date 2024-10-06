package test_test

import (
	"os"
	"regexp"
	"strings"
	"sync/atomic"
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

type TestParamMap map[string]TestParam

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

var testParams = TestParamMap{
	"base nothing": {
		test:   func(test.Test) {},
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
		test:     func(test.Test) { panic("fail") },
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
			func(test.Test) { panic("fail") }),
		expect:   test.Failure,
		consumed: true,
	},

	"inrun failure": {
		test: test.InRun(test.Failure,
			func(test.Test) {}),
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
			func(test.Test) { panic("fail") }),
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
		assert.True(t, finished)
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
			assert.True(t, finished)
		})
	}
}

func TestNewRunNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).Run(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			assert.Equal(t, tname, t.Name())
			testFailures(t, param)
		}).Cleanup(func() {
			assert.True(t, finished, tname)
		})
	}
}

func TestNewRunSeqNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).RunSeq(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			assert.Equal(t, tname, t.Name())
			testFailures(t, param)
		}).Cleanup(func() {
			assert.True(t, finished, tname)
		})
	}
}

func TestNewRunFiltered(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		pattern, finished := "base", false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.New[TestParam](t, TestParam{
			name:   test.Name(name),
			test:   param.test,
			expect: param.expect,
		}).Filter(pattern, true).Run(func(t test.Test, param TestParam) {
			defer func() { finished = true }()
			assert.Equal(t, tname, t.Name())
			assert.Contains(t, t.Name(), pattern)
			testFailures(t, param)
		}).Cleanup(func() {
			if strings.Contains(tname, pattern) {
				assert.True(t, finished, tname)
			}
		})
	}
}

func TestMapRun(t *testing.T) {
	count := atomic.Int32{}

	test.Map(t, testParams).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			testFailures(t, param)
		}).
		Cleanup(func() {
			assert.Equal(t, len(testParams), int(count.Load()))
		})
}

func TestMapRunFiltered(t *testing.T) {
	pattern, count := "base", atomic.Int32{}
	expect := testParams.FilterBy(pattern)

	test.Map(t, testParams).Filter(pattern, true).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			assert.Contains(t, t.Name(), pattern)
			testFailures(t, param)
		}).
		Cleanup(func() {
			assert.Equal(t, len(expect), int(count.Load()))
		})
}

func TestSliceRun(t *testing.T) {
	count := atomic.Int32{}

	test.Slice(t, testParams.GetSlice()).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			testFailures(t, param)
		}).
		Cleanup(func() {
			assert.Equal(t, len(testParams), int(count.Load()))
		})
}

func TestSliceRunFiltered(t *testing.T) {
	pattern, count := "inrun", atomic.Int32{}
	expect := testParams.FilterBy(pattern)

	test.Slice(t, testParams.GetSlice()).Filter(pattern, true).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			assert.Contains(t, t.Name(), pattern)
			testFailures(t, param)
		}).
		Cleanup(func() {
			assert.Equal(t, len(expect), int(count.Load()))
		})
}

type ParamParam struct {
	name   string
	expect bool
}

func TestTempDir(t *testing.T) {
	test.New[ParamParam](t, ParamParam{expect: true}).
		Run(func(t test.Test, _ ParamParam) {
			assert.NotEmpty(t, t.TempDir())
		})
}

func TestNameCastFallback(t *testing.T) {
	test.New[ParamParam](t, ParamParam{name: "value"}).
		Run(func(t test.Test, _ ParamParam) {
			assert.Equal(t, t.Name(), "TestNameCastFallback")
		})
}

func TestExpectCastFallback(t *testing.T) {
	test.New[ParamParam](t, ParamParam{expect: false}).
		Run(func(test.Test, ParamParam) {})
}

func TestTypePanic(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "not paniced")
		}
	}()
	test.New[TestParam](t, ParamParam{expect: false}).
		Run(func(test.Test, TestParam) {})
}

type PanicParam struct {
	parallel bool
	before   func(test.Test)
	during   func(test.Test)
	expect   mock.SetupFunc
}

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
		expect: test.Panic("testing: t.Setenv called after t.Parallel;" +
			" cannot set environment variables in parallel tests"),
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
		expect: test.Panic("testing: t.Parallel called after t.Setenv;" +
			" cannot set environment variables in parallel tests"),
	},

	"swallow multiple parallel calls": {
		during: func(t test.Test) {
			t.Parallel()
			t.Parallel()
		},
	},
}

func TestTesterPanic(t *testing.T) {
	for name, param := range testPanicParams {
		name, param := name, param
		t.Run(name, test.RunSeq(test.Success, func(t test.Test) {
			// Given
			if param.before != nil {
				mock.NewMocks(t).Expect(param.expect)
				param.before(t)
			}

			// When
			test.NewTester(t, test.Success).Run(func(t test.Test) {
				mock.NewMocks(t).Expect(param.expect)
				param.during(t)
			}, param.parallel)
		}))
	}
}

// This test is checking the runner for recovering from panics in parallel
// tests. Currently, I have no idea hot to integrate the test using the above
// simplified test pattern that only works on `test.Test` and not `testing.T“.
func TestRunnerPanic(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			assert.Fail(t, "not paniced")
		} else if v.(string) != "testing: t.Parallel called after t.Setenv;"+
			" cannot set environment variables in parallel tests" {
			assert.Fail(t, "unexpected panic: %v", v)
		}
	}()
	t.Setenv("TESTING", "before")

	test.New[ParamParam](t, []ParamParam{{expect: true}}).
		Run(func(_ test.Test, _ ParamParam) {})
}
