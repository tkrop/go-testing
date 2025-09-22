package test_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/test"
)

// TestParamRun is testing the test runner with single test cases.
func TestParamRun(t *testing.T) {
	finished := false
	test.Param(t, TestParam{
		test:   func(t test.Test) { t.FailNow() },
		expect: test.Failure,
	}).Run(func(t test.Test, param TestParam) {
		defer func() { finished = true }()
		param.CheckName(t)
		param.ExecTest(t)
	}).Cleanup(func() {
		assert.True(t, finished)
	})
}

// TestParamRunSeq is testing the test runner with single test cases running in
// sequence.
func TestParamRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		test.Param(t, param.Rename(name)).
			RunSeq(func(t test.Test, param TestParam) {
				defer func() { finished = true }()
				param.CheckName(t)
				param.ExecTest(t)
			}).
			Cleanup(func() {
				assert.True(t, finished)
			})
	}
}

// TestParamRunNamed is testing the test runner with single named test cases.
func TestParamRunNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			Run(func(t test.Test, param TestParam) {
				defer func() { finished = true }()
				assert.Equal(t, tname, t.Name())
				param.CheckName(t)
				param.ExecTest(t)
			}).
			Cleanup(func() {
				assert.True(t, finished, tname)
			})
	}
}

// TestParamRunSeqNamed is testing the test runner with single named test cases
// running in sequence.
func TestParamRunSeqNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			RunSeq(func(t test.Test, param TestParam) {
				defer func() { finished = true }()
				assert.Equal(t, tname, t.Name())
				param.CheckName(t)
				param.ExecTest(t)
			}).
			Cleanup(func() {
				assert.True(t, finished, tname)
			})
	}
}

// TestParamRunFiltered is testing the test runner with single named test cases
// using run while applying a filter.
func TestParamRunFiltered(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		pattern, finished := "base", false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			Filter(pattern, true).
			Run(func(t test.Test, param TestParam) {
				defer func() { finished = true }()
				assert.Equal(t, tname, t.Name())
				assert.Contains(t, t.Name(), pattern)
				assert.Contains(t, string(param.name), pattern)
				param.CheckName(t)
				param.ExecTest(t)
			}).
			Cleanup(func() {
				if strings.Contains(tname, pattern) {
					assert.True(t, finished, tname)
				}
			})
	}
}

// TestParamsRun is testing the test runner with parameterized tests.
func TestParamsRun(t *testing.T) {
	count := atomic.Int32{}

	test.Param(t, testParams.GetSlice()...).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(testParams), int(count.Load()))
		})
}

// TestParamsRunFiltered is testing the test runner with parameterized tests
// while applying a filter.
func TestParamsRunFiltered(t *testing.T) {
	pattern, count := "inrun", atomic.Int32{}
	expect := testParams.FilterBy(pattern)

	test.Param(t, testParams.GetSlice()...).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			name := string(param.name)
			assert.Contains(t, name, pattern)
			assert.NotNil(t, expect[name])
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(expect), int(count.Load()))
		})
}

// TestMapRun is testing the test runner with maps.
func TestMapRun(t *testing.T) {
	count := atomic.Int32{}

	test.Map(t, testParams).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(testParams), int(count.Load()))
		})
}

// TestMapRunFiltered is testing the test runner with maps while applying a
// filter.
func TestMapRunFiltered(t *testing.T) {
	pattern, count := "base", atomic.Int32{}
	expect := testParams.FilterBy(pattern)

	test.Map(t, testParams).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			assert.Contains(t, t.Name(), pattern)
			name := strings.ReplaceAll(t.Name()[19:], "-", " ")
			assert.Contains(t, name, pattern)
			assert.NotNil(t, expect[name])
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(expect), int(count.Load()))
		})
}

// TestSliceRun is testing the test runner with slices.
func TestSliceRun(t *testing.T) {
	count := atomic.Int32{}

	test.Slice(t, testParams.GetSlice()).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(testParams), int(count.Load()))
		})
}

// TestSliceRunFiltered is testing the test runner with slices while applying
// a filter.
func TestSliceRunFiltered(t *testing.T) {
	pattern, count := "inrun", atomic.Int32{}
	expect := testParams.FilterBy(pattern)

	test.Slice(t, testParams.GetSlice()).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParam) {
			defer count.Add(1)
			name := string(param.name)
			assert.Contains(t, name, pattern)
			assert.NotNil(t, expect[name])
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(expect), int(count.Load()))
		})
}

// This test is checking the runner for recovering from panics in parallel
// tests. Currently, I have no idea hot to integrate the test using the above
// simplified test pattern that only works on `test.Test` and not `testing.T“.
func TestRunnerPanic(t *testing.T) {
	defer func() {
		assert.Equal(t, "testing: test using t.Setenv or t.Chdir"+
			" can not use t.Parallel", recover())
	}()
	t.Setenv("TESTING", "before")

	test.Any[ParamParam](t, []ParamParam{{expect: true}}).
		Run(func(t test.Test, param ParamParam) {
			param.CheckName(t)
		})
}

// This test is checking the runner for recovering from panics in parallel
// tests. Currently, I have no idea hot to integrate the test using the above
// simplified test pattern that only works on `test.Test` and not `testing.T“.
func TestInvalidTypePanic(t *testing.T) {
	defer func() {
		assert.Equal(t, test.NewErrInvalidType(ParamParam{}), recover())
	}()

	test.Any[TestParam](t, ParamParam{expect: false}).
		Run(func(t test.Test, param TestParam) {
			param.CheckName(t)
		})
}

func TestNameCastFallback(t *testing.T) {
	test.Param(t, ParamParam{name: "value"}).
		Run(func(t test.Test, _ ParamParam) {
			assert.Equal(t, t.Name(), "TestNameCastFallback")
		})
}

func TestExpectCastFallback(t *testing.T) {
	test.Param(t, ParamParam{expect: false}).
		Run(func(t test.Test, param ParamParam) {
			param.CheckName(t)
		})
}
