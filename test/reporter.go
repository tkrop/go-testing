package test

import (
	"fmt"

	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/mock"
)

// Reporter is a minimal interface for abstracting test report methods that are
// needed to setup an isolated test environment for GoMock and Testify.
type Reporter interface {
	// Error reports a failure messages when a test is supposed to continue.
	Error(args ...any)
	// Errorf reports a failure messages when a test is supposed to continue.
	Errorf(format string, args ...any)
	// Fatal reports a fatal failure message that immediate aborts of the test
	// execution.
	Fatal(args ...any)
	// Fatalf reports a fatal failure message that immediate aborts of the test
	// execution.
	Fatalf(format string, args ...any)
	// Fail reports a failure message that immediate aborts of the test
	// execution.
	Fail()
	// FailNow reports fatal failure notifications without log output that
	// aborts test execution immediately.
	FailNow()
	// Panic reports a panic.
	Panic(arg any)
}

// Validator a test failure validator based on the test reporter interface.
type Validator struct {
	ctrl     *gomock.Controller
	recorder *Recorder
}

// Recorder a test failure validator recorder.
type Recorder struct {
	validator *Validator
}

// NewValidator creates a new test validator for validating error messages and
// panics created during test execution.
func NewValidator(ctrl *gomock.Controller) *Validator {
	validator := &Validator{ctrl: ctrl}
	validator.recorder = &Recorder{validator: validator}
	if t, ok := ctrl.T.(*Context); ok {
		// We need to install a second isolated test environment to break the
		// reporter cycle on the failure issued by the mock controller.
		ctrl.T = New(t.t, t.expect, t.parallel)
		t.expect = Failure
		t.Reporter(validator)
	}
	return validator
}

// EXPECT implements the usual `gomock.EXPECT` call to request the recorder.
func (v *Validator) EXPECT() *Recorder {
	return v.recorder
}

// Error receive expected method call to `Error`.
func (v *Validator) Error(args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Error", args...)
}

// Error indicate an expected method call to `Error`.
func (r *Recorder) Error(args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Error",
		reflect.TypeOf((*Validator)(nil).Error), args...)
}

// Error creates a validation method call setup for `Error`.
func Error(args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Error(args...).Do(mocks.Do(Reporter.Error))
	}
}

// Errorf receive expected method call to `Errorf`.
func (v *Validator) Errorf(format string, args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Errorf", append([]any{format}, args...)...)
}

// Errorf indicate an expected method call to `Errorf`.
func (r *Recorder) Errorf(format string, args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Errorf",
		reflect.TypeOf((*Validator)(nil).Errorf),
		append([]any{format}, args...)...)
}

// Errorf creates a validation method call setup for `Errorf`.
func Errorf(format string, args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Errorf(format, args...).Do(mocks.Do(Reporter.Errorf))
	}
}

// Fatal receive expected method call to `Fatal`.
func (v *Validator) Fatal(args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Fatal", args...)
}

// Fatal indicate an expected method call to `Fatal`.
func (r *Recorder) Fatal(args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Fatal",
		reflect.TypeOf((*Validator)(nil).Fatal), args...)
}

// Fatal creates a validation method call setup for `Fatal`.
func Fatal(args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Fatal(args...).Do(mocks.Do(Reporter.Fatal))
	}
}

// Fatalf receive expected method call to `Fatalf`.
func (v *Validator) Fatalf(format string, args ...any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Fatalf", append([]any{format}, args...)...)
}

// Fatalf indicate an expected method call to `Fatalf`.
func (r *Recorder) Fatalf(format string, args ...any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Fatalf",
		reflect.TypeOf((*Validator)(nil).Fatalf),
		append([]any{format}, args...)...)
}

// Fatalf creates a validation method call setup for `Fatalf`.
func Fatalf(format string, args ...any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Fatalf(format, args...).Do(mocks.Do(Reporter.Fatalf))
	}
}

// Fail receive expected method call to `Fail`.
func (v *Validator) Fail() {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Fail")
}

// Fail indicate an expected method call to `Fail`.
func (r *Recorder) Fail() *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Fail",
		reflect.TypeOf((*Validator)(nil).Fail))
}

// Fail creates a validation method call setup for `Fail`.
func Fail() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Fail().Do(mocks.Do(Reporter.Fail))
	}
}

// FailNow receive expected method call to `FailNow`.
func (v *Validator) FailNow() {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "FailNow")
}

// FailNow indicate an expected method call to `FailNow`.
func (r *Recorder) FailNow() *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "FailNow",
		reflect.TypeOf((*Validator)(nil).FailNow))
}

// FailNow creates a validation method call setup for `FailNow`.
func FailNow() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			FailNow().Do(mocks.Do(Reporter.FailNow))
	}
}

// Panic receive expected method call indicating a panic.
func (v *Validator) Panic(arg any) {
	v.ctrl.T.Helper()
	v.ctrl.Call(v, "Panic", []any{arg}...)
}

// Panic indicate an expected method call from panic.
func (r *Recorder) Panic(arg any) *gomock.Call {
	r.validator.ctrl.T.Helper()
	return r.validator.ctrl.RecordCallWithMethodType(r.validator, "Panic",
		reflect.TypeOf((*Validator)(nil).Panic), []any{arg}...)
}

// Panic creates a validation method call setup for a panic. It allows to match
// the panic object, which usually is an error and alternatively the error
// string representing the error, since runtime errors may be irreproducible.
func Panic(arg any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewValidator).EXPECT().
			Panic(EqError(arg)).Do(mocks.Do(Reporter.Panic))
	}
}

// UnexpectedCall creates expectation for unexpected calls. We only support one
// unexpected call since the test execution stops in this case.
func UnexpectedCall[T any](
	creator func(*gomock.Controller) *T,
	method, caller string, args ...any,
) func(Test, *mock.Mocks) mock.SetupFunc {
	return func(_ Test, mocks *mock.Mocks) mock.SetupFunc {
		return Fatalf("Unexpected call to %T.%v(%v) at %s because: %s",
			mock.Get(mocks, creator), method, reflect.StringArgs(args), caller,
			//nolint:err113 // necessary
			fmt.Errorf("there are no expected calls "+
				"of the method \"%s\" for that receiver", method))
	}
}

func ConsumedCall[T any](
	creator func(*gomock.Controller) *T,
	method, caller, ecaller string, args ...any,
) func(Test, *mock.Mocks) mock.SetupFunc {
	return func(_ Test, mocks *mock.Mocks) mock.SetupFunc {
		return Fatalf("Unexpected call to %T.%v(%v) at %s because: %s",
			mock.Get(mocks, creator), method, reflect.StringArgs(args), caller,
			fmt.Errorf("\nexpected call at %s has "+ //nolint:err113 // necessary
				"already been called the max number of times", ecaller))
	}
}

// MissingCalls creates an expectation for all missing calls.
func MissingCalls(
	setups ...mock.SetupFunc,
) func(Test, *mock.Mocks) mock.SetupFunc {
	return func(t Test, _ *mock.Mocks) mock.SetupFunc {
		// Creates a new mock controller and test environment to isolate the
		// validator used for sub-call creation/registration from the validator
		// used for execution.
		mocks := mock.NewMocks(New(t, false, false))
		calls := make([]func(*mock.Mocks) any, 0, len(setups))
		for _, setup := range setups {
			calls = append(calls,
				Errorf("missing call(s) to %v", EqCall(setup(mocks))))
		}
		calls = append(calls, Errorf("aborting test due to missing call(s)"))
		return mock.Chain(calls...)
	}
}

// errorMatcher is a matcher to improve capabilities of matching errors.
type errorMatcher struct {
	x any
}

// EqError creates a new error matcher that allows to match either the error or
// alternatively the string describing the error.
func EqError(x any) gomock.Matcher {
	return &errorMatcher{x: x}
}

// Matches executes the extended error matching.
func (m *errorMatcher) Matches(x any) bool {
	switch a := m.x.(type) {
	case string:
		switch b := x.(type) {
		case string:
			return a == b
		case error:
			return a == b.Error()
		}
	case error:
		switch b := x.(type) {
		case string:
			return a.Error() == b
		case error:
			return gomock.Eq(a).Matches(b)
		}
	}
	return gomock.Eq(m.x).Matches(x)
}

// String creates a string of the expectation to match.
func (m *errorMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", m.x, m.x)
}

// callMatcher is a matcher that supports matching of calls. Calls contain
// actions consisting of functions that cannot be matched successfully using
// [reflect.DeepEquals].
type callMatcher struct {
	x any
}

// EqCall creates a new call matcher that allows to match calls by translating
// them to the string containing the core information instead of using the
// standard matcher using [reflect.DeepEquals] that fails for the contained
// actions.
func EqCall(x any) gomock.Matcher {
	return &callMatcher{x: x}
}

// Matches executes the extended call matching algorithms.
func (m *callMatcher) Matches(x any) bool {
	a, aok := m.x.(*gomock.Call)
	if b, bok := x.(*gomock.Call); aok && bok {
		return a.String() == b.String()
	}
	return gomock.Eq(m.x).Matches(x)
}

// String creates a string of the expectation to match.
func (m *callMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", m.x, m.x)
}
