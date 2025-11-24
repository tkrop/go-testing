package test

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/internal/slices"
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

// FilterFunc defines the common test filter function signature.
type FilterFunc[P any] func(name string, param P) bool

// CleanupFunc defines the common test cleanup function signature.
type CleanupFunc func()

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

// Not negates the given filter function.
func Not[P any](filter FilterFunc[P]) FilterFunc[P] {
	return func(name string, param P) bool {
		return !filter(name, param)
	}
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
	// Adds a generic filter function that allows to filter test cases based on
	// the name and the parameter set.
	Filter(filter FilterFunc[P]) Factory[P]
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
	// Cleanup register a function to be called to cleanup after all tests have
	// finished to remove the shared resources.
	Cleanup(call CleanupFunc)
}

// factory is a generic parameterized test factory struct.
type factory[P any] struct {
	// The testing context to run the tests in.
	t *testing.T
	// A wait group to synchronize the test execution.
	wg sync.WaitGroup
	// The test parameter sets to run.
	params any
	// A filter to include or exclude test cases.
	filter []func(string, P) bool
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
func Any[P any](t *testing.T, params any) Factory[P] {
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
func Param[P any](t *testing.T, params ...P) Factory[P] {
	t.Helper()

	if len(params) == 1 {
		return Any[P](t, params[0])
	}
	return Any[P](t, params)
}

// Map creates a new parallel test runner with given test parameter sets
// provided as a test case name to parameter sets mapping.
func Map[P any](t *testing.T, params ...map[string]P) Factory[P] {
	t.Helper()

	return Any[P](t, maps.Add(maps.Copy(params[0]), params[1:]...))
}

// Slice creates a new parallel test runner with given test parameter sets
// provided as a slice. The test runner is looking into the parameter set to
// find a suitable test case name.
func Slice[P any](t *testing.T, params ...[]P) Factory[P] {
	t.Helper()

	return Any[P](t, slices.Add(params...))
}

// Filter adds a generic filter function that allows to filter test cases based
// on the name and the parameter set.
func (r *factory[P]) Filter(filter FilterFunc[P]) Factory[P] {
	r.filter = append(r.filter, filter)
	return r
}

// Timeout can be used to set up a timeout for the test cases executed by the
// test runner. Setting a timeout is useful to prevent the test execution from
// waiting too long in case of deadlocks. The timeout is not affecting the
// global test timeout that may only abort a test earlier. If the given
// duration is zero or negative, the timeout is ignored.
func (r *factory[P]) Timeout(timeout time.Duration) Factory[P] {
	r.timeout = timeout
	return r
}

// StopEarly can be used to stop the test by the given duration ahead of an
// individual or global test deadline. This is useful to ensure that resources
// can be cleaned up before the global deadline is exceeded.
func (r *factory[P]) StopEarly(early time.Duration) Factory[P] {
	r.early = early
	return r
}

// Run runs the test parameter sets (by default) parallel.
func (r *factory[P]) Run(call ParamFunc[P]) Factory[P] {
	return r.run(call, Parallel)
}

// RunSeq runs the test parameter sets in a sequence.
func (r *factory[P]) RunSeq(call ParamFunc[P]) Factory[P] {
	return r.run(call, !Parallel)
}

// Cleanup register a function to be called for cleanup after all tests have
// been finished - successful and failing.
func (r *factory[P]) Cleanup(call CleanupFunc) {
	r.t.Cleanup(func() {
		r.t.Helper()
		r.wg.Wait()
		call()
	})
}

// Parallel ensures that the test runner runs the test parameter sets in
// parallel.
func (r *factory[P]) parallel(parallel bool) {
	if parallel {
		defer r.recover()
		r.t.Parallel()
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

// Runs the test parameter sets defined by the factory either parallel or in
// sequence.
func (r *factory[P]) run(
	call ParamFunc[P], parallel bool,
) Factory[P] {
	switch params := r.params.(type) {
	case map[string]P:
		r.parallel(parallel)
		for name, param := range params {
			name := reflect.Name(name, param)
			r.exec(name, param, call, parallel)
		}

	case []P:
		r.parallel(parallel)
		for index, param := range params {
			name := reflect.Name("", param) + "[" + strconv.Itoa(index) + "]"
			r.exec(name, param, call, parallel)
		}

	case P:
		name := reflect.Name("", params)
		r.exec(name, params, call, parallel)
	default:
		panic(NewErrInvalidType(r.params))
	}
	return r
}

// Executes the given test parameter set with the provided name after matching
// against the filters. If one of the applied filters matches the test case it
// is skipped.
func (r *factory[P]) exec(
	name string, param P, call ParamFunc[P], parallel bool,
) {
	for _, filter := range r.filter {
		if !filter(name, param) {
			return
		}
	}

	// Execute anonymous non-parallel tests directly.
	if name == "" && !parallel {
		r.wrap(param, call, parallel)(r.t)
		return
	}

	// TODO: Think about how https://pkg.go.dev/testing/synctes can be utilized
	// here to better coordinate parallel tests.
	r.t.Run(name, r.wrap(param, call, parallel))
}

// wrap creates the wrapper method eventually executing the test.
func (r *factory[P]) wrap(
	param P, call ParamFunc[P], parallel bool,
) func(*testing.T) {
	r.wg.Add(1)

	return func(t *testing.T) {
		t.Helper()

		New(t, parallel).
			Expect(reflect.Find(param, Success, "expect", "*")).
			Timeout(reflect.Find(param, r.timeout, "timeout")).
			StopEarly(reflect.Find(param, r.early, "early")).
			Run(func(t Test) {
				t.Helper()

				defer r.wg.Done()
				call(t, param)
			})
	}
}
