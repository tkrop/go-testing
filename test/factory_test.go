package test_test

import (
	"maps"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/iter"
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

// This test is checking the runner for rerunnercovering from panics in parallel
// tests. Currently, I have no idea hot to integrate the test using the above
// simplified test pattern that only works on `test.Test` and not `testing.T“.
func TestRunnerPanic(t *testing.T) {
	defer test.Recover(t, "testing: test using t.Setenv, t.Chdir, or "+
		"cryptotest.SetGlobalRandom can not use t.Parallel")
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
	// all filter - always includes all cases
	"all-single-case": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.All[Any]())
		},
		expect: map[string]bool{
			"test-case": true,
		},
	},

	"all-multiple-cases": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.All[Any]())
		},
		expect: map[string]bool{
			"alpha-test": true,
			"beta-test":  true,
			"gamma-test": true,
		},
	},

	"all-with-and": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.And(
				test.All[Any](),
				test.Pattern[Any]("beta"),
			))
		},
		expect: map[string]bool{
			"beta-test": true,
		},
	},

	"all-with-or": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Or(
				test.All[Any](),
				test.Pattern[Any]("beta"),
			))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"beta-test":  true,
			"gamma":      true,
		},
	},

	// none filter - always excludes all cases
	"none-single-case": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.None[Any]())
		},
		expect: map[string]bool{},
	},

	"none-multiple-cases": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma-test": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.None[Any]())
		},
		expect: map[string]bool{},
	},

	"none-with-and": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.And(
				test.None[Any](),
				test.Pattern[Any]("beta"),
			))
		},
		expect: map[string]bool{},
	},

	"none-with-or": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Or(
				test.None[Any](),
				test.Pattern[Any]("beta"),
			))
		},
		expect: map[string]bool{
			"beta-test": true,
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

	// Tests with xor - exactly one filter must match
	"xor-none-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Xor(
				test.Pattern[Any]("delta"),
				test.Pattern[Any]("epsilon"),
			))
		},
		expect: map[string]bool{},
	},

	"xor-single-match": {
		cases: map[string]Any{
			"alpha": {},
			"beta":  {},
			"gamma": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Xor(
				test.Pattern[Any]("alpha"),
				test.Pattern[Any]("gamma"),
			))
		},
		expect: map[string]bool{
			"alpha": true,
			"gamma": true,
		},
	},

	"xor-double-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Xor(
				test.Pattern[Any]("alpha"),
				test.Pattern[Any]("test"),
			))
		},
		expect: map[string]bool{
			"beta-test": true,
		},
	},

	// Tests with implies - condition implies consequence
	"implies-false-condition": {
		cases: map[string]Any{
			"alpha": {},
			"beta":  {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.Pattern[Any]("delta"),
				test.Pattern[Any]("alpha"),
			))
		},
		expect: map[string]bool{
			"alpha": true,
			"beta":  true,
		},
	},

	"implies-true-condition-true-consequence": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.Pattern[Any]("test"),
				test.Pattern[Any]("alpha"),
			))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"gamma":      true,
		},
	},

	"implies-true-condition-false-consequence": {
		cases: map[string]Any{
			"alpha": {},
			"beta":  {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.OS[Any](runtime.GOOS),
				test.Pattern[Any]("never-matches"),
			))
		},
		expect: map[string]bool{},
	},

	// Tests with implies - multiple consequences (all must hold)
	"implies-multi-all-consequences-match": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.Pattern[Any]("test"),
				test.Pattern[Any]("alpha"),
				test.Pattern[Any]("test"),
			))
		},
		expect: map[string]bool{
			"alpha-test": true,
			"gamma":      true,
		},
	},

	"implies-multi-first-consequence-fails": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.Pattern[Any]("test"),
				test.Pattern[Any]("never-matches"),
				test.Pattern[Any]("alpha"),
			))
		},
		expect: map[string]bool{
			"gamma": true,
		},
	},

	"implies-multi-last-consequence-fails": {
		cases: map[string]Any{
			"alpha-test": {},
			"beta-test":  {},
			"gamma":      {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Filter(test.Implies(
				test.Pattern[Any]("test"),
				test.Pattern[Any]("alpha"),
				test.Pattern[Any]("never-matches"),
			))
		},
		expect: map[string]bool{
			"gamma": true,
		},
	},

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
}

func TestFilter(t *testing.T) {
	t.Parallel()
	test.Map(t, filterCases).
		Run(func(t test.Test, param filterParams) {
			// Given
			expect := &sync.Map{}

			// When
			param.apply(test.Map(t, param.cases)).
				Prefix("any/prefix=").
				Run(func(t test.Test, _ Any) {
					expect.Store(strings.Split(t.Name(), "=")[1], true)
				}).
				Cleanup(func() {
					// Then
					assert.Equal(t, param.expect,
						maps.Collect(iter.SyncMap[string, bool](expect)))
				})
		})
}

type prefixParams struct {
	cases  map[string]Any
	apply  func(FactoryAny) FactoryAny
	expect map[string]bool
}

var prefixCases = map[string]prefixParams{
	"no-prefix": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny { return f },
		expect: map[string]bool{
			"test-case": true,
		},
	},
	"single-prefix": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Prefix("my-prefix/")
		},
		expect: map[string]bool{
			"my-prefix/test-case": true,
		},
	},

	"prefix-last": {
		cases: map[string]Any{
			"test-case": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Prefix("group-a/").Prefix("group-b/")
		},
		expect: map[string]bool{
			"group-b/test-case": true,
		},
	},

	// Multiple cases with prefix
	"prefix-multiple-cases": {
		cases: map[string]Any{
			"alpha": {},
			"beta":  {},
			"gamma": {},
		},
		apply: func(f FactoryAny) FactoryAny {
			return f.Prefix("group/")
		},
		expect: map[string]bool{
			"group/alpha": true,
			"group/beta":  true,
			"group/gamma": true,
		},
	},
}

func TestPrefix(t *testing.T) {
	test.Map(t, prefixCases).
		Run(func(t test.Test, param prefixParams) {
			// Given
			parent := t.Name()
			expect := &sync.Map{}

			// When
			param.apply(test.Map(t, param.cases)).
				RunSeq(func(t test.Test, _ Any) {
					t.Parallel()

					rel := strings.TrimPrefix(t.Name(), parent+"/")
					expect.Store(rel, true)
				}).
				Cleanup(func() {
					// Then
					assert.Equal(t, param.expect,
						maps.Collect(iter.SyncMap[string, bool](expect)))
				})
		})
}
