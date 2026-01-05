package test_test

import (
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/test"
)

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
	pattern, count := "inrun-failure", atomic.Int32{}
	expect := commonTestCases.FilterBy(pattern)
	assert.NotEmpty(t, expect)

	test.Param(t, commonTestCases.GetSlice()...).
		Filter(test.Pattern[TestParams](pattern)).
		Run(func(t test.Test, param TestParams) {
			defer count.Add(1)
			name := param.name
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
		Filter(test.Pattern[TestParams](pattern)).
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
	pattern, count := "inrun-success", atomic.Int32{}
	expect := commonTestCases.FilterBy(pattern)
	assert.NotEmpty(t, expect)

	test.Slice(t, commonTestCases.GetSlice()).
		Filter(test.Pattern[TestParams](pattern)).
		Run(func(t test.Test, param TestParams) {
			defer count.Add(1)
			name := param.name
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
	defer test.Recover(t, "testing: test using t.Setenv or t.Chdir"+
		" can not use t.Parallel")
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
	defer test.Recover(t, test.NewErrInvalidType(ParamParams{}))

	test.Any[TestParams](t, ParamParams{expect: false}).
		Run(func(t test.Test, param TestParams) {
			param.CheckName(t)
		})
}

type (
	Any          = struct{}
	FactoryAny   = test.Factory[Any]
	filterParams struct {
		cases  map[string]Any
		apply  func(FactoryAny) FactoryAny
		expect map[string]bool
	}
)

var filterCases = map[string]filterParams{
	// Test filter by name variants
	"include-single-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Pattern[Any]("beta"))
		},
		expect: map[string]bool{
			"beta-test": true,
		},
	},

	"exclude-single-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Not(test.Pattern[Any]("beta")))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"gamma":      true,
		},
	},

	"include-space-normalized": {
		cases: map[string]Any{
			"alpha-test":      {},
			"beta-test":       {},
			"beta-test-extra": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Pattern[Any]("beta test"))
		},
		expect: map[string]bool{
			"beta-test":       true,
			"beta-test-extra": true,
		},
	},

	"exclude-space-normalized": {
		cases: map[string]Any{
			"alpha-test":      {},
			"beta-test":       {},
			"beta-test-extra": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Not(test.Pattern[Any]("beta test")))
		},
		expect: map[string]bool{
			"alpha-test": true,
		},
	},

	"include-no-matches": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Pattern[Any]("delta"))
		},
		expect: map[string]bool{},
	},

	"exclude-all-matches": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Not(test.Pattern[Any]("test")))
		},
		expect: map[string]bool{},
	},

	// Test with pattern/os/arch combined
	"pattern-filter-include": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Pattern[Any]("beta"))
		},
		expect: map[string]bool{
			"beta-test": true,
		},
	},

	"pattern-filter-exclude-with-spaces": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(func(name string, _ Any) bool {
				return !strings.Contains(name, "beta")
			})
		},
		expect: map[string]bool{
			"alpha-test": true,
			"gamma":      true,
		},
	},

	"os-filter-current-os": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.OS[Any](runtime.GOOS))
		},
		expect: map[string]bool{
			"test-case": true,
		},
	},

	"os-filter-different-os": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f test.Factory[Any]) test.Factory[Any] {
			other := "linux"
			if runtime.GOOS == "linux" {
				other = "windows"
			}
			return f.Filter(test.OS[Any](other))
		},
		expect: map[string]bool{},
	},

	"arch-filter-current-arch": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Arch[Any](runtime.GOARCH))
		},
		expect: map[string]bool{
			"test-case": true,
		},
	},

	"arch-filter-different-arch": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f test.Factory[Any]) test.Factory[Any] {
			other := "amd64"
			if runtime.GOARCH == "amd64" {
				other = "arm64"
			}
			return f.Filter(test.Arch[Any](other))
		},
		expect: map[string]bool{},
	},

	"multiple-filters-combined": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
		},
		apply: func(f test.Factory[Any]) test.Factory[Any] {
			return f.
				Filter(test.Pattern[Any]("test")).
				Filter(test.OS[Any](runtime.GOOS)).
				Filter(func(name string, _ Any) bool {
					return strings.Contains(name, "beta") ||
						strings.Contains(name, "gamma")
				})
		},
		expect: map[string]bool{
			"beta-test":  true,
			"gamma-test": true,
		},
	},

	// Tests with and/or combined
	"and-filter-all-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.And(
				test.Pattern[Any]("test"),
				func(name string, _ Any) bool {
					return strings.Contains(name, "beta")
				},
			))
		},
		expect: map[string]bool{
			"beta-test": true,
		},
	},

	"and-filter-no-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.And(
				test.Pattern[Any]("test"),
				func(name string, _ Any) bool {
					return strings.Contains(name, "delta")
				},
			))
		},
		expect: map[string]bool{},
	},

	"or-filter-single-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Or(
				func(name string, _ Any) bool {
					return strings.Contains(name, "alpha")
				},
				func(name string, _ Any) bool {
					return strings.Contains(name, "gamma")
				},
			))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"gamma":      true,
		},
	},

	"or-filter-multiple-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Or(
				func(name string, _ Any) bool {
					return strings.Contains(name, "alpha")
				},
				func(name string, _ Any) bool {
					return strings.Contains(name, "beta")
				},
			))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"beta-test":  true,
		},
	},

	"or-filter-no-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Or(
				func(name string, _ Any) bool {
					return strings.Contains(name, "delta")
				},
				func(name string, _ Any) bool {
					return strings.Contains(name, "epsilon")
				},
			))
		},
		expect: map[string]bool{},
	},

	"and-or-combined": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
			"delta":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.And(
				test.Pattern[Any]("test"),
				test.Or(
					func(name string, _ Any) bool {
						return strings.Contains(name, "beta")
					},
					func(name string, _ Any) bool {
						return strings.Contains(name, "gamma")
					},
				),
			))
		},
		expect: map[string]bool{
			"beta-test":  true,
			"gamma-test": true,
		},
	},
}

func TestFilter(t *testing.T) {
	t.Parallel()

	for name, param := range filterCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			expect := map[string]bool{}

			// When
			param.apply(test.Map(t, param.cases)).
				RunSeq(func(t test.Test, _ Any) {
					parts := strings.Split(t.Name(), "/")
					expect[parts[len(parts)-1]] = true
				}).
				Cleanup(func() {
					// Then
					assert.Equal(t, param.expect, expect)
				})
		})
	}
}
