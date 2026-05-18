package test_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// TODO: check whether test validation is strong enough to detect failures.

type benchParams struct {
	value int
}

var benchTestCases = map[string]benchParams{
	"case-one": {value: 1},
	"case-two": {value: 2},
}

func TestBenchmarkRun(t *testing.T) {
	// Given
	var seen sync.Map

	// When
	testing.Benchmark(func(b *testing.B) {
		b.Helper()

		test.Map(test.Benchmark(b), benchTestCases).
			Run(func(_ test.Test, param benchParams) {
				seen.Store(param.value, true)
			})
	})

	// Then
	var got int32
	seen.Range(func(_, _ any) bool { got++; return true })
	assert.Equal(t, int32(len(benchTestCases)), got)
}

type wrapperParams struct {
	call   func(t test.Test) any
	expect any
}

var wrapperTestCases = map[string]wrapperParams{
	// Parallel must not panic and must be a no-op.
	"parallel-noop": {
		call: func(t test.Test) any {
			t.Parallel()
			return nil
		},
	},

	// Deadline must return the zero time and false.
	"deadline-zero": {
		call: func(t test.Test) any {
			deadline, ok := t.Deadline()
			return [2]any{deadline, ok}
		},
		expect: [2]any{time.Time{}, false},
	},
}

func TestBenchmarkWrapper(t *testing.T) {
	test.Map(t, wrapperTestCases).
		Run(func(t test.Test, param wrapperParams) {
			// Given
			var got any

			// When
			testing.Benchmark(func(b *testing.B) {
				b.Helper()

				got = param.call(test.Benchmark(b))
			})

			// Then
			assert.Equal(t, param.expect, got)
		})
}

type dispatchParams struct {
	setup   mock.SetupFunc
	factory func(t test.Test) test.Factory[benchParams]
	expect  int32
}

var dispatchTestCases = map[string]dispatchParams{
	// Map dispatches all cases.
	"map-all": {
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Map(t, benchTestCases)
		},
		expect: int32(len(benchTestCases)),
	},

	// Map with filter dispatches only matching cases.
	"map-filtered": {
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Map(t, benchTestCases).
				Filter(test.Pattern[benchParams]("case-one"))
		},
		expect: 1,
	},

	// Param dispatches all variadic cases.
	"param-all": {
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Param(t,
				benchTestCases["case-one"],
				benchTestCases["case-two"])
		},
		expect: int32(len(benchTestCases)),
	},

	// Slice dispatches all slice cases.
	"slice-all": {
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Slice(t, []benchParams{
				benchTestCases["case-one"],
				benchTestCases["case-two"],
			})
		},
		expect: int32(len(benchTestCases)),
	},

	// Any with a single value dispatches exactly one case.
	"any-single": {
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Any[benchParams](t, benchTestCases["case-one"])
		},
		expect: 1,
	},

	// non-benchmarker panics when Benchmark is called.
	"non-benchmarker": {
		setup: test.Panic(
			"not a benchmark target [use test.Benchmark(b)]"),
		factory: func(t test.Test) test.Factory[benchParams] {
			return test.Map(t, benchTestCases)
		},
	},
}

func TestBenchmarkDispatch(t *testing.T) {
	test.Map(t, dispatchTestCases).
		Run(func(t test.Test, param dispatchParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)
			var seen sync.Map
			call := func(_ *testing.B, p benchParams) func(*testing.B) {
				seen.Store(p.value, true)
				return func(*testing.B) {}
			}

			// When
			if param.setup == nil {
				testing.Benchmark(func(b *testing.B) {
					b.Helper()

					param.factory(test.Benchmark(b)).Benchmark(call)
				})
			} else {
				param.factory(t).Benchmark(call)
			}

			// Then
			var got int32
			seen.Range(func(_, _ any) bool { got++; return true })
			assert.Equal(t, param.expect, got)
		})
}
