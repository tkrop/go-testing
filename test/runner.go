package test

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/internal/sync"
)

// ErrInvalidType is an error for invalid types.
var ErrInvalidType = errors.New("invalid type")

// NewErrInvalidType creates a new invalid type error.
func NewErrInvalidType(value any) error {
	return fmt.Errorf("%w [type: %v]",
		ErrInvalidType, reflect.ValueOf(value).Type())
}

// TestName returns the normalized test case name for the given name and given
// parameter set. If the name is empty, the name is resolved from the parameter
// set using the `name` parameter. The resolved name is normalized before being
// returned.
func TestName[P any](name string, param P) string {
	if name != "" {
		return strings.ReplaceAll(name, " ", "-")
	} else if name := Find(param, unknown, "name", "*"); name != "" {
		return strings.ReplaceAll(string(name), " ", "-")
	}
	return string(unknown)
}

// SetupFunc defines the common test setup function signature.
type SetupFunc func(Test)

// ParamFunc defines the common parameterized test function signature.
type ParamFunc[P any] func(t Test, param P)

// CleanupFunc defines the common test cleanup function signature.
type CleanupFunc func()

// Runner is a generic test runner interface.
type Runner[P any] interface {
	// Filter sets up a filter for the test cases using the given pattern and
	// match flag. The pattern is a regular expression that is matched against
	// the test case name. The match flag is used to include or exclude the
	// test cases that match the pattern.
	Filter(pattern string, match bool) Runner[P]
	// Timeout sets up a timeout for the test cases executed by the test runner.
	// Setting a timeout is useful to prevent the test execution from waiting
	// too long in case of deadlocks. The timeout is not affecting the global
	// test timeout that may only abort a test earlier. If the given duration is
	// zero or negative, the timeout is ignored.
	Timeout(timeout time.Duration) Runner[P]
	// StopEarly stops the test by the given duration ahead of an individual or
	// global test deadline. This is useful to ensure that resources can be
	// cleaned up before the global deadline is exceeded.
	StopEarly(time time.Duration) Runner[P]
	// Run runs all test parameter sets in parallel. If the test parameter sets
	// are provided as a map, the test case name is used as the test name. If
	// the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	Run(call ParamFunc[P]) Runner[P]
	// RunSeq runs the test parameter sets in a sequence. If the test parameter
	// sets are provided as a map, the test case name is used as the test name.
	// If the test parameter sets are provided as a slice, the test case name is
	// created by appending the index to the test name. If the test parameter
	// sets are provided as a single parameter set, the test case name is used
	// as the test name. The test case name is normalized before being used.
	RunSeq(call ParamFunc[P]) Runner[P]
	// Cleanup register a function to be called to cleanup after all tests have
	// finished to remove the shared resources.
	Cleanup(call CleanupFunc)
}

// runner is a generic parameterized test runner struct.
type runner[P any] struct {
	// The testing context to run the tests in.
	t *testing.T
	// A wait group to synchronize the test execution.
	wg sync.WaitGroup
	// The test parameter sets to run.
	params any
	// A filter to include or exclude test cases.
	filter func(string) bool
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
func Any[P any](t *testing.T, params any) Runner[P] {
	t.Helper()

	return &runner[P]{
		t:      t,
		wg:     sync.NewWaitGroup(),
		params: params,
	}
}

// Map creates a new parallel test runner with given test parameter sets
// provided as a test case name to parameter sets mapping.
func Map[P any](t *testing.T, params ...map[string]P) Runner[P] {
	t.Helper()

	return Any[P](t, maps.Add(maps.Copy(params[0]), params[1:]...))
}

// Slice creates a new parallel test runner with given test parameter sets
// provided as a slice. The test runner is looking into the parameter set to
// find a suitable test case name.
func Slice[P any](t *testing.T, params []P) Runner[P] {
	t.Helper()

	return Any[P](t, params)
}

// Filter filters the test cases by the given pattern and match flag. The
// pattern is a regular expression that is matched against the test case name.
// The match flag is used to include or exclude the test cases that match the
// pattern.
func (r *runner[P]) Filter(
	pattern string, match bool,
) Runner[P] {
	regexp := regexp.MustCompile(pattern)
	r.filter = func(name string) bool {
		return regexp.MatchString(name) == match
	}
	return r
}

// Timeout can be used to set up a timeout for the test cases executed by the
// test runner. Setting a timeout is useful to prevent the test execution from
// waiting too long in case of deadlocks. The timeout is not affecting the
// global test timeout that may only abort a test earlier. If the given
// duration is zero or negative, the timeout is ignored.
func (r *runner[P]) Timeout(timeout time.Duration) Runner[P] {
	r.timeout = timeout
	return r
}

// StopEarly can be used to stop the test by the given duration ahead of an
// individual or global test deadline. This is useful to ensure that resources
// can be cleaned up before the global deadline is exceeded.
func (r *runner[P]) StopEarly(early time.Duration) Runner[P] {
	r.early = early
	return r
}

// Run runs the test parameter sets (by default) parallel.
func (r *runner[P]) Run(call ParamFunc[P]) Runner[P] {
	return r.run(call, Parallel)
}

// RunSeq runs the test parameter sets in a sequence.
func (r *runner[P]) RunSeq(call ParamFunc[P]) Runner[P] {
	return r.run(call, !Parallel)
}

// Cleanup register a function to be called for cleanup after all tests have
// been finished - successful and failing.
func (r *runner[P]) Cleanup(call CleanupFunc) {
	r.t.Cleanup(func() {
		r.t.Helper()
		r.wg.Wait()
		call()
	})
}

// Parallel ensures that the test runner runs the test parameter sets in
// parallel.
func (r *runner[P]) parallel(parallel bool) {
	if parallel {
		defer r.recoverParallel()
		r.t.Parallel()
	}
}

// RecoverParallel recovers from panics when calling `t.Parallel()` multiple
// times.
func (*runner[P]) recoverParallel() {
	//revive:disable-next-line:defer // only used inside a deferred call.
	if v := recover(); v != nil &&
		v != "testing: t.Parallel called multiple times" {
		panic(v)
	}
}

// Run runs the test parameter sets either parallel or in sequence.
func (r *runner[P]) run(
	call ParamFunc[P], parallel bool,
) Runner[P] {
	switch params := r.params.(type) {
	case map[string]P:
		r.parallel(parallel)
		for name, param := range params {
			name := TestName(name, param)
			if r.filter != nil && !r.filter(name) {
				continue
			}
			r.t.Run(name, r.setup(param, call, parallel))
		}

	case []P:
		r.parallel(parallel)
		for index, param := range params {
			name := TestName("", param) + "[" + strconv.Itoa(index) + "]"
			if r.filter != nil && !r.filter(name) {
				continue
			}
			r.t.Run(name, r.setup(param, call, parallel))
		}

	case P:
		name := TestName("", params)
		if r.filter != nil && !r.filter(name) {
			return r
		}
		if name != string(unknown) {
			r.t.Run(name, r.setup(params, call, parallel))
		} else {
			r.setup(params, call, parallel)(r.t)
		}

	default:
		panic(NewErrInvalidType(r.params))
	}
	return r
}

// setup sets up the test case by creating the wrapper method, registering the
// test case in the waiting group, calling the test specific setup function and
// preparing the test specific cleanup function, iff the setup and the cleanup
// functions are defined by the parameter sett and provide non-nil values.
func (r *runner[P]) setup(
	param P, call ParamFunc[P], parallel bool,
) func(*testing.T) {
	r.wg.Add(1)

	access := NewAccessor(param)
	if access != nil {
		var setup SetupFunc
		before := access.Find(setup, "before")
		if before, ok := before.(SetupFunc); ok && before != nil {
			before(r.t)
		}

		var cleanup CleanupFunc
		after := access.Find(cleanup, "after")
		if after, ok := after.(CleanupFunc); ok && after != nil {
			r.t.Cleanup(after)
		}
	}
	return r.test(param, call, parallel)
}

// test creates the wrapper method executing eventually the test.
func (r *runner[P]) test(
	param P, call ParamFunc[P], parallel bool,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		New(t, Find(param, Success, "expect", "*")).
			Timeout(Find(param, r.timeout, "timeout")).
			StopEarly(Find(param, r.early, "early")).
			Run(func(t Test) {
				t.Helper()

				defer r.wg.Done()
				call(t, param)
			}, parallel)
	}
}
