package test

import (
	"fmt"

	"github.com/golang/mock/gomock"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/mock"
)

// Validator a test failure validator based on the thest reporter interface.
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
	if t, ok := ctrl.T.(*Tester); ok {
		// We need to install a second isolated test environment to break the
		// reporter cycle on the failure issued by the mock controller.
		ctrl.T = NewTester(t.t, t.expect)
		t.expect = Failure
		t.Reporter(validator)
	}
	return validator
}

// EXPECT implements the usual `gomock.EXPECT` call to request the recorder.
func (v *Validator) EXPECT() *Recorder {
	return v.recorder
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
			Errorf(format, args...).
			Do(mocks.Return(Reporter.Errorf))
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
			Fatalf(format, args...).Do(mocks.Return(Reporter.Fatalf))
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
			FailNow().Do(mocks.Return(Reporter.FailNow))
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
			Panic(Error(arg)).Do(mocks.Return(Reporter.Panic))
	}
}

// errorMatcher is an error matcher to improve capabilities of matching errors.
type errorMatcher struct {
	x any
}

// Error creates a new error matcher that allows to match either the error or
// alternatively the string describing the error.
func Error(x any) gomock.Matcher {
	return &errorMatcher{x: x}
}

// Matches executes the extended error matching.
func (m *errorMatcher) Matches(x any) bool {
	switch a := m.x.(type) {
	case string:
		switch b := x.(type) {
		case string:
			return gomock.Eq(a).Matches(b)
		case error:
			return gomock.Eq(a).Matches(b.Error())
		}
	case error:
		switch b := x.(type) {
		case string:
			return gomock.Eq(a.Error()).Matches(b)
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
