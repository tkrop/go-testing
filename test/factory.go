package test

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	imaps "github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/internal/sync"
	"github.com/tkrop/go-testing/reflect"
)

// ErrInvalidType is an error for invalid types.
var ErrInvalidType = errors.New("invalid type")

// NewErrInvalidType creates a new invalid type error.
func NewErrInvalidType(value any) error {
	return fmt.Errorf("%w [type: %T]", ErrInvalidType, value)
}

// SetupFunc defines the common test setup function signature.
type SetupFunc func(Test)

// ParamFunc defines the common parameterized test function signature.
type ParamFunc[P any] func(t Test, param P)

// BenchmarkFunc defines a two-phase parameterized benchmark function
// signature. The signature requires a setup function that returns a loop
// function to be executed in the benchmark loop. The setup phase receives the
// raw `*testing.B` and parameters for benchmark configuration, and returns
// a loop function. The benchmark automatically calls `b.ReportAllocs` and
// `b.ResetTimer` after the setup phase and before entering the loop.
type BenchmarkFunc[P any] func(b *testing.B, param P) func(*testing.B)

// FilterFunc defines the common test filter function signature expecting the
// test name and parameter set.
type FilterFunc[P any] func(name string, param P) bool

// CleanupFunc defines the common test cleanup function signature.
type CleanupFunc func()

// All creates a filter function that always returns true and thereby filters
// all test cases.
func All[P any]() FilterFunc[P] {
	return func(_ string, _ P) bool {
		return true
	}
}

// None creates a filter function that always returns false and thereby filters
// no test cases.
func None[P any]() FilterFunc[P] {
	return func(_ string, _ P) bool {
		return false
	}
}

// Not negates the given filter function.
func Not[P any](filter FilterFunc[P]) FilterFunc[P] {
	return func(name string, param P) bool {
		return !filter(name, param)
	}
}

// And combines the given filter functions with a logical `and`.
func And[P any](filters ...FilterFunc[P]) FilterFunc[P] {
	return func(name string, param P) bool {
		for _, filter := range filters {
			if !filter(name, param) {
				return false
			}
		}
		return true
	}
}

// Or combines the given filter functions with a logical `or`.
func Or[P any](filters ...FilterFunc[P]) FilterFunc[P] {
	return func(name string, param P) bool {
		for _, filter := range filters {
			if filter(name, param) {
				return true
			}
		}
		return false
	}
}

// Xor combines the given filter functions with an exclusive or, returning
// true if exactly one of the filters matches.
func Xor[P any](filters ...FilterFunc[P]) FilterFunc[P] {
	return func(name string, param P) bool {
		count := 0
		for _, filter := range filters {
			if filter(name, param) {
				count++
			}
		}
		return count == 1
	}
}

// Implies is a convenience filter implementing logical implication. It is
// returning true if the condition does not match or all consequences match.
func Implies[P any](
	when FilterFunc[P], then ...FilterFunc[P],
) FilterFunc[P] {
	return Or(Not(when), And(then...))
}

// Pattern creates a filter function that matches the test case name against the
// given pattern. The pattern is adjusted to replace spaces with dashes before
// being compiled into a regular expression to account for the test name
// normalization.
func Pattern[P any](pattern string) FilterFunc[P] {
	pattern = strings.ReplaceAll(pattern, " ", "-")
	regexp := regexp.MustCompile(pattern)
	return func(name string, _ P) bool {
		return regexp.MatchString(name)
	}
}

// OS creates a filter function that matches the given operating system.
func OS[P any](os string) FilterFunc[P] {
	return func(_ string, _ P) bool {
		return runtime.GOOS == os
	}
}

// Arch creates a filter function that matches the given architecture.
func Arch[P any](arch string) FilterFunc[P] {
	return func(_ string, _ P) bool {
		return runtime.GOARCH == arch
	}
}

// Factory is a generic test factory interface.
type Factory[P any] interface {
	// Prefix sets a test and benchmark prefix with given value. This can be
	// used to structure tests and benchmarks into logical groups add labels
	// for tools. If called multiple times, only the last prefix is applied.
	Prefix(prefix string) Factory[P]
	// Adds generic filter functions that allows to filter test cases based on
	// the name and the parameter set. If multiple filters are added, all of
	// them must accept the test case, else the test case is excluded. Thus the
	// filters are combined by default using a logical `and`.
	//
	// **Note:** A test case prefix is always applied to the test case name
	// before the test case parameters and name are passed to the filter.
	Filter(filter ...FilterFunc[P]) Factory[P]
	// Timeout sets up a timeout for the test cases executed by the test runner.
	// Setting a timeout is useful to prevent the test execution from waiting
	// too long in case of deadlocks. The timeout is not affecting the global
	// test timeout that may only abort a test earlier. If the given duration is
	// zero or negative, the timeout is ignored.
	Timeout(timeout time.Duration) Factory[P]
	// StopEarly stops the test by the given duration ahead of an individual or
	// global test deadline. This is useful to ensure that resources can be
	// cleaned up before the global deadline is exceeded.
	StopEarly(time time.Duration) Factory[P]
	// Run runs all test parameter sets in parallel. If the test parameter sets
	// are provided as a map, the test case name is used as the test name. If
	// the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	Run(call ParamFunc[P]) Factory[P]
	// RunSeq runs the test parameter sets in a sequence. If the test parameter
	// sets are provided as a map, the test case name is used as the test name.
	// If the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	RunSeq(call ParamFunc[P]) Factory[P]
	// Benchmark runs all test parameter sets using the two-phase benchmarker
	// pattern. The function is called for setup and returns the loop function
	// to be executed in the benchmark loop. The benchmark automatically calls
	// `b.ReportAllocs` and `b.ResetTimer` after the setup phase and before
	// entering the loop.
	Benchmark(call BenchmarkFunc[P]) Factory[P]
	// Cleanup register a function to be called to cleanup after all tests have
	// finished to remove the shared resources.
	Cleanup(call CleanupFunc)
}

// factory is a generic parameterized test factory struct.
type factory[P any] struct {
	// The testing context to run the tests in.
	t Test
	// A wait group to synchronize the test execution.
	wg sync.WaitGroup
	// The test parameter sets to run.
	params any
	// A prefix to prepend to each test case name.
	prefix string
	// A filters to include or exclude test cases.
	filters []FilterFunc[P]
	// A timeout after which the test execution is stopped to prevent waiting
	// to long in case of deadlocks.
	timeout time.Duration
	// A time reserved for cleaning up resources before reaching the deadline.
	early time.Duration
}

// Any creates a new parallel test runner with given parameter set(s). The set
// can be a single test parameter set, a slice of test parameter sets, or a map
// of named test parameter sets. The test runner is looking into the parameter
// set to determine a suitable test case name, e.g. by using a `name` parameter.
func Any[P any](t Test, params any) Factory[P] {
	t.Helper()

	return &factory[P]{
		t:      t,
		wg:     sync.NewWaitGroup(),
		params: params,
	}
}

// Param creates a new parallel test runner with given test parameter sets
// provided as variadic arguments. The test runner is looking into the
// parameter set to find a suitable test case name.
func Param[P any](t Test, params ...P) Factory[P] {
	t.Helper()

	if len(params) == 1 {
		return Any[P](t, params[0])
	}
	return Any[P](t, params)
}

// Map creates a new parallel test runner with given test parameter sets
// provided as a test case name to parameter sets mapping.
//
// Note: The test cases are sorted by their lexicographical order to ensure a
// deterministic, stable test order.
func Map[P any](t Test, params ...map[string]P) Factory[P] {
	t.Helper()

	return Any[P](t, imaps.Copy(maps.Clone(params[0]), params[1:]...))
}

// Slice creates a new parallel test runner with given test parameter sets
// provided as a slice. The test runner is looking into the parameter set to
// find a suitable test case name.
func Slice[P any](t Test, params ...[]P) Factory[P] {
	t.Helper()

	return Any[P](t, slices.Concat(params...))
}

// Filter adds a generic filter function that allows to filter test cases based
// on the name and the parameter set.
func (f *factory[P]) Filter(filter ...FilterFunc[P]) Factory[P] {
	f.filters = append(f.filters, filter...)
	return f
}

// Prefix sets a test and benchmark prefix with the given value. This can be
// used to structure tests and benchmarks into logical groups or add labels.
// If called multiple times, only the last prefix is applied.
func (f *factory[P]) Prefix(prefix string) Factory[P] {
	f.prefix = prefix
	return f
}

// Timeout can be used to set up a timeout for the test cases executed by the
// test runner. Setting a timeout is useful to prevent the test execution from
// waiting too long in case of deadlocks. The timeout is not affecting the
// global test timeout that may only abort a test earlier. If the given
// duration is zero or negative, the timeout is ignored.
func (f *factory[P]) Timeout(timeout time.Duration) Factory[P] {
	f.timeout = timeout
	return f
}

// StopEarly can be used to stop the test by the given duration ahead of an
// individual or global test deadline. This is useful to ensure that resources
// can be cleaned up before the global deadline is exceeded.
func (f *factory[P]) StopEarly(early time.Duration) Factory[P] {
	f.early = early
	return f
}

// Run runs the test parameter sets (by default) parallel.
func (f *factory[P]) Run(call ParamFunc[P]) Factory[P] {
	return f.dispatch(func(name string, param P) {
		f.test(name, param, call, Parallel)
	}, Parallel)
}

// RunSeq runs the test parameter sets in a sequence.
func (f *factory[P]) RunSeq(call ParamFunc[P]) Factory[P] {
	return f.dispatch(func(name string, param P) {
		f.test(name, param, call, Sequential)
	}, Sequential)
}

// Benchmark runs all parameter sets using the two-phase benchmark pattern.
// This is only supported for benchmarks and will panic if the test target does
// not implement the Benchmarker interface.
//
// Use `test.Benchmark(b)` to wrap the *testing.B target as a Benchmarker when
// using the test runner in benchmarks.
func (f *factory[P]) Benchmark(call BenchmarkFunc[P]) Factory[P] {
	return f.dispatch(func(name string, param P) {
		f.bench(name, param, call)
	}, Sequential)
}

// Cleanup register a function to be called for cleanup after all tests have
// been finished - successful and failing.
func (f *factory[P]) Cleanup(call CleanupFunc) {
	f.t.Cleanup(func() {
		f.t.Helper()
		f.wg.Wait()
		call()
	})
}

// Mode ensures that the test runner runs the test parameter sets in
// the specified mode.
func (f *factory[P]) mode(mode Mode) {
	if mode&Parallel == Parallel {
		defer f.recover()
		f.t.Parallel()
	}
}

// Recover recovers from panics when calling `t.Parallel()` multiple times.
func (*factory[P]) recover() {
	//revive:disable-next-line:defer // only used inside a deferred call.
	if v := recover(); v != nil &&
		v != "testing: t.Parallel called multiple times" {
		panic(v)
	}
}

// dispatch iterates over the parameter sets, resolves names, applies filters,
// and calls the provided function for each accepted `(name, param)` pair. For
// map and slice cases it also triggers the outer parallel declaration if
// enabled.
func (f *factory[P]) dispatch(
	call func(name string, param P), mode Mode,
) Factory[P] {
	switch params := f.params.(type) {
	case map[string]P:
		f.mode(mode)
		for _, key := range f.sorted(params) {
			param := params[key]
			name := reflect.Name(key, param)
			f.filter(name, param, call)
		}

	case []P:
		f.mode(mode)
		for index, param := range params {
			name := reflect.Name("", param) +
				"[" + strconv.Itoa(index) + "]"
			f.filter(name, param, call)
		}

	case P:
		name := reflect.Name("", params)
		f.filter(name, params, call)

	default:
		panic(NewErrInvalidType(f.params))
	}

	return f
}

// sorted sorts the keys of the test case map by the lexicographical order of
// the test case name including the prefix to ensure a deterministic, stable
// test order independent of the natural map iteration order.
func (f *factory[P]) sorted(params map[string]P) []string {
	return slices.SortedFunc(maps.Keys(params), func(a, b string) int {
		return strings.Compare(
			f.prefix+reflect.Name(a, params[a]),
			f.prefix+reflect.Name(b, params[b]),
		)
	})
}

// filter filters the parameter set by its `(name, param)` pair before calling
// the provided function. The function is only called when all registered
// filters accept the test case.
//
// *Note:* The name is prefixed if setup before being passed to the filters.
func (f *factory[P]) filter(
	name string, param P, call func(name string, param P),
) {
	if f.prefix != "" {
		name = f.prefix + name
	}

	for _, filter := range f.filters {
		if !filter(name, param) {
			return
		}
	}

	call(name, param)
}

// wrap creates the wrapper method eventually executing the test.
func (f *factory[P]) wrap(
	name string, param P, call ParamFunc[P], mode Mode,
) func(Test) {
	f.wg.Add(1)

	return func(t Test) {
		t.Helper()

		New(t).Mode(mode).
			Expect(reflect.Find(param, Success, "expect", "*")).
			Timeout(reflect.Find(param, f.timeout, "timeout")).
			StopEarly(reflect.Find(param, f.early, "early")).
			Run(name, func(t Test) {
				t.Helper()

				defer f.wg.Done()
				call(t, param)
			})
	}
}

// test executes the given single test parameter set with the given name. If
// the test is configured to run in parallel, it is executed in a separate
// goroutine and synchronized with the wait group. The test is wrapped to
// apply the timeout and early stop configuration.
func (f *factory[P]) test(
	name string, param P, call ParamFunc[P], mode Mode,
) {
	// Execute anonymous non-parallel tests directly.
	if name == "" && mode&Parallel == Sequential {
		f.wrap("", param, call, mode)(f.t)
		return
	}

	New(f.t).Mode(Sequential).Run(name,
		f.wrap("", param, call, mode))
}

// bench runs the given single bench parameter set with the given name.
func (f *factory[P]) bench(
	name string, param P, call BenchmarkFunc[P],
) {
	b, ok := f.t.(Benchmarker)
	if !ok {
		panic("not a benchmark target [use test.Benchmark(b)]")
	}
	f.wg.Add(1)
	defer f.wg.Done()

	b.Benchmark(name, func(b *testing.B) {
		b.Helper()

		call := call(b, param)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			call(b)
		}
	})
}
