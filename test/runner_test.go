package test_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/test"
)

// TestAnyRun is testing the test runner with single test cases.
func TestAnyRun(t *testing.T) {
	finished := false
	test.Any[TestParam](t, TestParam{
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

// TestAnyRunSeq is testing the test runner with single test cases running in
// sequence.
func TestAnyRunSeq(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		test.Any[TestParam](t, param.Rename(name)).
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

// TestAnyRunNamed is testing the test runner with single named test cases.
func TestAnyRunNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Any[TestParam](t, param.Rename(name)).
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

// TestAnyRunSeqNamed is testing the test runner with single named test cases
// running in sequence.
func TestAnyRunSeqNamed(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		finished := false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Any[TestParam](t, param.Rename(name)).
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

// TestAnyRunFiltered is testing the test runner with single named test cases
// using run while applying a filter.
func TestAnyRunFiltered(t *testing.T) {
	t.Parallel()

	for name, param := range testParams {
		pattern, finished := "base", false
		tname := t.Name() + "/" + test.TestName(name, param)
		test.Any[TestParam](t, param.Rename(name)).
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
	test.Any[ParamParam](t, ParamParam{name: "value"}).
		Run(func(t test.Test, _ ParamParam) {
			assert.Equal(t, t.Name(), "TestNameCastFallback")
		})
}

func TestExpectCastFallback(t *testing.T) {
	test.Any[ParamParam](t, ParamParam{expect: false}).
		Run(func(t test.Test, param ParamParam) {
			param.CheckName(t)
		})
}
