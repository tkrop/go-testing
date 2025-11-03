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
	test.Param(t, TestParams{
		test:   func(t test.Test) { t.FailNow() },
		expect: test.Failure,
	}).Run(func(t test.Test, param TestParams) {
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

	for name, param := range commonTestCases {
		finished := false
		test.Param(t, param.Rename(name)).
			RunSeq(func(t test.Test, param TestParams) {
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

	for name, param := range commonTestCases {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			Run(func(t test.Test, param TestParams) {
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

	for name, param := range commonTestCases {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			RunSeq(func(t test.Test, param TestParams) {
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

	for name, param := range commonTestCases {
		pattern, finished := "base", false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Param(t, param.Rename(name)).
			Filter(pattern, true).
			Run(func(t test.Test, param TestParams) {
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

	test.Param(t, commonTestCases.GetSlice()...).
		Run(func(t test.Test, param TestParams) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(commonTestCases), int(count.Load()))
		})
}

// TestParamsRunFiltered is testing the test runner with parameterized tests
// while applying a filter.
func TestParamsRunFiltered(t *testing.T) {
	pattern, count := "inrun failure", atomic.Int32{}
	expect := commonTestCases.FilterBy(pattern)
	assert.NotEmpty(t, expect)

	test.Param(t, commonTestCases.GetSlice()...).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParams) {
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

	test.Map(t, commonTestCases).
		Run(func(t test.Test, param TestParams) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(commonTestCases), int(count.Load()))
		})
}

// TestMapRunFiltered is testing the test runner with maps while applying a
// filter.
func TestMapRunFiltered(t *testing.T) {
	pattern, count := "base", atomic.Int32{}
	expect := commonTestCases.FilterBy(pattern)
	assert.NotEmpty(t, expect)

	test.Map(t, commonTestCases).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParams) {
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

	test.Slice(t, commonTestCases.GetSlice()).
		Run(func(t test.Test, param TestParams) {
			defer count.Add(1)
			param.CheckName(t)
			param.ExecTest(t)
		}).
		Cleanup(func() {
			assert.Equal(t, len(commonTestCases), int(count.Load()))
		})
}

// TestSliceRunFiltered is testing the test runner with slices while applying
// a filter.
func TestSliceRunFiltered(t *testing.T) {
	pattern, count := "inrun success", atomic.Int32{}
	expect := commonTestCases.FilterBy(pattern)
	assert.NotEmpty(t, expect)

	test.Slice(t, commonTestCases.GetSlice()).
		Filter(pattern, true).
		Run(func(t test.Test, param TestParams) {
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

	test.Any[ParamParams](t, []ParamParams{{expect: true}}).
		Run(func(t test.Test, param ParamParams) {
			param.CheckName(t)
		})
}

// This test is checking the runner for recovering from panics in parallel
// tests. Currently, I have no idea hot to integrate the test using the above
// simplified test pattern that only works on `test.Test` and not `testing.T“.
func TestInvalidTypePanic(t *testing.T) {
	defer func() {
		assert.Equal(t, test.NewErrInvalidType(ParamParams{}), recover())
	}()

	test.Any[TestParams](t, ParamParams{expect: false}).
		Run(func(t test.Test, param TestParams) {
			param.CheckName(t)
		})
}

func TestNameCastFallback(t *testing.T) {
	test.Param(t, ParamParams{name: "value"}).
		Run(func(t test.Test, _ ParamParams) {
			assert.Equal(t, t.Name(), "TestNameCastFallback")
		})
}

func TestExpectCastFallback(t *testing.T) {
	test.Param(t, ParamParams{expect: false}).
		Run(func(t test.Test, param ParamParams) {
			param.CheckName(t)
		})
}
