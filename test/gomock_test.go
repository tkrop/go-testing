package test_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

const (
	anyError   = "any error"
	otherError = "other error"
)

var (
	errAny   = errors.New(anyError)
	errOther = errors.New(otherError)
)

type ErrorMatcherParams struct {
	base   any
	match  any
	result bool
}

var testErrorMatcherParams = map[string]ErrorMatcherParams{
	"success-string-string": {
		base:   anyError,
		match:  anyError,
		result: true,
	},
	"success-string-error": {
		base:   anyError,
		match:  errAny,
		result: true,
	},
	"success-error-string": {
		base:   errAny,
		match:  anyError,
		result: true,
	},
	"success-error-error": {
		base:   errAny,
		match:  errAny,
		result: true,
	},
	"success-other-other": {
		base:   1,
		match:  1,
		result: true,
	},

	"failure-string-string": {
		base:   errAny,
		match:  errOther,
		result: false,
	},
	"failure-string-error": {
		base:   anyError,
		match:  errOther,
		result: false,
	},
	"failure-error-string": {
		base:   errAny,
		match:  otherError,
		result: false,
	},
	"failure-error-error": {
		base:   errAny,
		match:  errOther,
		result: false,
	},
	"failure-other-other": {
		base:   1,
		match:  false,
		result: false,
	},
}

func TestErrorMatcher(t *testing.T) {
	test.Map(t, testErrorMatcherParams).
		Run(func(t test.Test, param ErrorMatcherParams) {
			// Given
			matcher := test.Error(param.base)

			// When
			result := matcher.Matches(param.match)

			// Then
			assert.Equal(t, param.result, result)
		})
}

func TestErrorMatcherString(t *testing.T) {
	assert.Equal(t, test.Error(true).String(), "is equal to true (bool)")
}

func testFatalf(
	method string, caller string, args ...any,
) func(mocks *mock.Mocks) mock.SetupFunc {
	return func(mocks *mock.Mocks) mock.SetupFunc {
		return test.Fatalf("Unexpected call to %T.%v(%v) at %s because: %s",
			mock.Get(mocks, test.NewValidator), method, args, caller,
			errors.New("there are no expected calls of the method \""+
				method+"\" for that receiver"))
	}
}

func testMissing(setup mock.SetupFunc) func(mocks *mock.Mocks) mock.SetupFunc {
	return func(mocks *mock.Mocks) mock.SetupFunc {
		return test.Errorf("missing call(s) to %v",
			setup(mocks).([]any)...)
	}
}

type ReporterParams struct {
	mockSetup mock.SetupFunc
	failSetup func(mocks *mock.Mocks) mock.SetupFunc
	call      func(t test.Test)
}

var testReporterParams = map[string]ReporterParams{
	"errorf called": {
		mockSetup: test.Errorf("fail"),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"fatalf called": {
		mockSetup: test.Fatalf("fail"),
		call: func(t test.Test) {
			t.Fatalf("fail")
		},
	},
	"failnow called": {
		mockSetup: test.FailNow(),
		call: func(t test.Test) {
			t.FailNow()
		},
	},
	"panic called": {
		mockSetup: test.Panic("fail"),
		call: func(t test.Test) {
			panic("fail")
		},
	},

	"errorf missing": {
		mockSetup: test.Errorf("fail"),
		failSetup: testMissing(test.Errorf("fail")),
		call:      func(t test.Test) {},
	},
	"fatalf missing": {
		mockSetup: test.Fatalf("fail"),
		failSetup: testMissing(test.Fatalf("fail")),
		call:      func(t test.Test) {},
	},
	"failnow missing": {
		mockSetup: test.FailNow(),
		failSetup: testMissing(test.FailNow()),
		call:      func(t test.Test) {},
	},
	"panic missing": {
		mockSetup: test.Panic("fail"),
		failSetup: testMissing(test.Panic("fail")),
		call:      func(t test.Test) {},
	},

	"errorf undeclared": {
		failSetup: testFatalf("Errorf", CallerErrorf, "fail"),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"fatalf undeclared": {
		failSetup: testFatalf("Fatalf", CallerFatalf, "fail"),
		call: func(t test.Test) {
			t.Fatalf("fail")
		},
	},
	"failnow undeclared": {
		failSetup: testFatalf("FailNow", CallerFailNow),
		call: func(t test.Test) {
			t.FailNow()
		},
	},
	"panic undeclared": {
		failSetup: testFatalf("Panic", CallerPanic, "fail"),
		call: func(t test.Test) {
			panic("fail")
		},
	},
}

func TestReporter(t *testing.T) {
	test.Map(t, testReporterParams).
		Run(func(t test.Test, param ReporterParams) {
			// Given
			mocks := mock.NewMocks(t)

			// When
			test.InRun(test.Failure, func(tt test.Test) {
				// Given
				xmocks := mock.NewMocks(tt)

				if param.failSetup != nil {
					mocks.Expect(param.failSetup(xmocks))
				}
				xmocks.Expect(param.mockSetup)

				// Connect the mock controller directly to the isolated parent
				// test environment to capture the mock controller failure.
				xmocks.Ctrl.T = t

				// When
				param.call(tt)
			})(t)
		})
}
