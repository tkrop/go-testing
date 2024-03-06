package test_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
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

type MatcherParams struct {
	matcher       func(any) gomock.Matcher
	base          any
	match         any
	expectMatches bool
	expectString  string
}

var testErrorMatcherParams = map[string]MatcherParams{
	"error-matcher-success-string-string": {
		matcher:       test.EqError,
		base:          anyError,
		match:         anyError,
		expectMatches: true,
		expectString:  "is equal to any error (string)",
	},
	"error-matcher-success-string-error": {
		matcher:       test.EqError,
		base:          anyError,
		match:         errAny,
		expectMatches: true,
		expectString:  "is equal to any error (string)",
	},
	"error-matcher-success-error-string": {
		matcher:       test.EqError,
		base:          errAny,
		match:         anyError,
		expectMatches: true,
		expectString:  "is equal to any error (*errors.errorString)",
	},
	"error-matcher-success-error-error": {
		matcher:       test.EqError,
		base:          errAny,
		match:         errAny,
		expectMatches: true,
		expectString:  "is equal to any error (*errors.errorString)",
	},
	"error-matcher-success-other-other": {
		matcher:       test.EqError,
		base:          1,
		match:         1,
		expectMatches: true,
		expectString:  "is equal to 1 (int)",
	},

	"error-matcher-failure-string-string": {
		matcher:       test.EqError,
		base:          errAny,
		match:         errOther,
		expectMatches: false,
		expectString:  "is equal to any error (*errors.errorString)",
	},
	"error-matcher-failure-string-error": {
		matcher:       test.EqError,
		base:          anyError,
		match:         errOther,
		expectMatches: false,
		expectString:  "is equal to any error (string)",
	},
	"error-matcher-failure-error-string": {
		matcher:       test.EqError,
		base:          errAny,
		match:         otherError,
		expectMatches: false,
		expectString:  "is equal to any error (*errors.errorString)",
	},
	"error-matcher-failure-error-error": {
		matcher:       test.EqError,
		base:          errAny,
		match:         errOther,
		expectMatches: false,
		expectString:  "is equal to any error (*errors.errorString)",
	},
	"error-matcher-failure-other-other": {
		matcher:       test.EqError,
		base:          1,
		match:         false,
		expectMatches: false,
		expectString:  "is equal to 1 (int)",
	},
}

func TestErrorMatcher(t *testing.T) {
	test.Map(t, testErrorMatcherParams).
		Run(func(t test.Test, param MatcherParams) {
			// Given
			matcher := param.matcher(param.base)

			// When
			matches := matcher.Matches(param.match)

			// Then
			assert.Equal(t, param.expectMatches, matches)
			assert.Equal(t, param.expectString, matcher.String())
		})
}

var testCallMatcherParams = map[string]MatcherParams{
	"call-matcher-success-call-call": {
		matcher:       test.EqCall,
		base:          test.Errorf("fail"),
		match:         test.Errorf("fail"),
		expectMatches: true,
		expectString: "is equal to *test.Validator.Errorf" +
			"(is equal to fail (string)) " + CallerGomockErrorf +
			" (*gomock.Call)",
	},
	"call-matcher-success-any-nay": {
		matcher:       test.EqCall,
		base:          "any string",
		match:         "any string",
		expectMatches: true,
		expectString:  "is equal to any string (string)",
	},
}

func evalCall(arg any, mocks *mock.Mocks) any {
	if call, ok := arg.(mock.SetupFunc); ok {
		return call(mocks)
	}
	return arg
}

func TestCallMatcher(t *testing.T) {
	test.Map(t, testCallMatcherParams).
		Run(func(t test.Test, param MatcherParams) {
			// Given - send mock calls to unchecked tester.
			mocks := mock.NewMocks(test.NewTester(t, test.Success))
			matcher := param.matcher(evalCall(param.base, mocks))

			// When
			matches := matcher.Matches(evalCall(param.match, mocks))

			// Then
			assert.Equal(t, param.expectMatches, matches)
			assert.Equal(t, param.expectString, matcher.String())
		})
}

type ReporterParams struct {
	mockSetup mock.SetupFunc
	failSetup func(test.Test, *mock.Mocks) mock.SetupFunc
	call      func(test.Test)
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
		call: func(test.Test) {
			panic("fail")
		},
	},

	"errorf undeclared": {
		failSetup: test.UnexpectedCall(test.NewValidator,
			"Errorf", CallerErrorf, "fail"),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"errorf undeclared twice": {
		failSetup: test.UnexpectedCall(test.NewValidator,
			"Errorf", CallerErrorf, "fail"),
		call: func(t test.Test) {
			t.Errorf("fail")
			t.Errorf("fail")
		},
	},
	"fatalf undeclared": {
		failSetup: test.UnexpectedCall(test.NewValidator,
			"Fatalf", CallerFatalf, "fail"),
		call: func(t test.Test) {
			t.Fatalf("fail")
		},
	},
	"failnow undeclared": {
		failSetup: test.UnexpectedCall(test.NewValidator,
			"FailNow", CallerFailNow),
		call: func(t test.Test) {
			t.FailNow()
		},
	},
	"panic undeclared": {
		failSetup: test.UnexpectedCall(test.NewValidator,
			"Panic", CallerPanic, "fail"),
		call: func(test.Test) {
			panic("fail")
		},
	},

	// Only Errorf can be consumed more than once, since Fatalf, FailNow, and
	// panic will stop execution immediately. The second call is effectively
	// unreachable.
	"errorf consumed": {
		mockSetup: test.Errorf("fail"),
		failSetup: test.ConsumedCall(test.NewValidator,
			"Errorf", CallerTestErrorf, CallerGomockErrorf, "fail"),
		call: func(t test.Test) {
			t.Errorf("fail")
			t.Errorf("fail")
		},
	},
	"fatalf consumed": {
		mockSetup: test.Fatalf("fail"),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing
			t.Fatalf("fail")
			t.Fatalf("fail")
		},
	},
	"failnow consumed": {
		mockSetup: test.FailNow(),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing
			t.FailNow()
			t.FailNow()
		},
	},
	"panic consumed": {
		mockSetup: test.Panic("fail"),
		call: func(test.Test) {
			panic("fail")
			//nolint:govet // needed for testing
			panic("fail")
		},
	},

	// The mock setup is automatically creating a [test.Validator] requiring
	// a the test environment to expect a failure to get called. To satisfy
	// this, we need to create at least one failure.
	"errorf missing": {
		mockSetup: mock.Chain(test.Errorf("fail"), test.Errorf("fail")),
		failSetup: test.MissingCalls(test.Errorf("fail")),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"errorf missing two calls": {
		mockSetup: mock.Chain(
			test.Errorf("fail"), test.Errorf("fail"), test.Errorf("fail-x"),
		),
		failSetup: test.MissingCalls(
			test.Errorf("fail"), test.Errorf("fail-x"),
		),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"fatalf missing": {
		mockSetup: mock.Chain(test.Errorf("fail"), test.Fatalf("fail")),
		failSetup: test.MissingCalls(test.Fatalf("fail")),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"failnow missing": {
		mockSetup: mock.Chain(test.Errorf("fail"), test.FailNow()),
		failSetup: test.MissingCalls(test.FailNow()),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
	"panic missing": {
		mockSetup: mock.Chain(test.Errorf("fail"), test.Panic("fail")),
		failSetup: test.MissingCalls(test.Panic("fail")),
		call: func(t test.Test) {
			t.Errorf("fail")
		},
	},
}

func TestReporter(t *testing.T) {
	test.Map(t, testReporterParams).
		Run(func(t test.Test, param ReporterParams) {
			// Given
			mocks := mock.NewMocks(t)

			// When
			test.InRun(test.Success, func(tt test.Test) {
				// Given
				imocks := mock.NewMocks(tt)
				if param.failSetup != nil {
					mocks.Expect(param.failSetup(tt, imocks))
				}
				imocks.Expect(param.mockSetup)

				// Connect the mock controller directly to the isolated parent
				// test environment to capture the mock controller failure.
				imocks.Ctrl.T = t

				// When
				param.call(tt)
			})(t)
		})
}
