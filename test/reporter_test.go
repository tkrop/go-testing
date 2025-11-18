package test_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

var (
	anErrorString  = assert.AnError.Error()
	anErrorMessage = "is equal to " + anErrorString + " "
	otherError     = "general other error for testing"
	errOther       = errors.New(otherError)
)

type MatcherParams struct {
	matcher       func(any) gomock.Matcher
	base          any
	match         any
	expectMatches bool
	expectString  string
}

var errorMatcherTestCases = map[string]MatcherParams{
	"success string-string": {
		matcher:       test.EqError,
		base:          anErrorString,
		match:         anErrorString,
		expectMatches: true,
		expectString:  anErrorMessage + "(string)",
	},
	"success string-error": {
		matcher:       test.EqError,
		base:          anErrorString,
		match:         assert.AnError,
		expectMatches: true,
		expectString:  anErrorMessage + "(string)",
	},
	"success error-string": {
		matcher:       test.EqError,
		base:          assert.AnError,
		match:         anErrorString,
		expectMatches: true,
		expectString:  anErrorMessage + "(*errors.errorString)",
	},
	"success error-error": {
		matcher:       test.EqError,
		base:          assert.AnError,
		match:         assert.AnError,
		expectMatches: true,
		expectString:  anErrorMessage + "(*errors.errorString)",
	},
	"success other-other": {
		matcher:       test.EqError,
		base:          1,
		match:         1,
		expectMatches: true,
		expectString:  "is equal to 1 (int)",
	},

	"failure string-string": {
		matcher:       test.EqError,
		base:          assert.AnError,
		match:         errOther,
		expectMatches: false,
		expectString:  anErrorMessage + "(*errors.errorString)",
	},
	"failure string-error": {
		matcher:       test.EqError,
		base:          anErrorString,
		match:         errOther,
		expectMatches: false,
		expectString:  anErrorMessage + "(string)",
	},
	"failure error-string": {
		matcher:       test.EqError,
		base:          assert.AnError,
		match:         otherError,
		expectMatches: false,
		expectString:  anErrorMessage + "(*errors.errorString)",
	},
	"failure error-error": {
		matcher:       test.EqError,
		base:          assert.AnError,
		match:         errOther,
		expectMatches: false,
		expectString:  anErrorMessage + "(*errors.errorString)",
	},
	"failure other-other": {
		matcher:       test.EqError,
		base:          1,
		match:         false,
		expectMatches: false,
		expectString:  "is equal to 1 (int)",
	},
}

func TestErrorMatcher(t *testing.T) {
	test.Map(t, errorMatcherTestCases).
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

var callMatcherTestCases = map[string]MatcherParams{
	"success-call-call": {
		matcher:       test.EqCall,
		base:          test.Errorf("%s", "fail"),
		match:         test.Errorf("%s", "fail"),
		expectMatches: true,
		expectString: "is equal to *test.Validator.Errorf" +
			"(is equal to %s (string), is equal to fail (string))" +
			" " + CallerReporterErrorf + " (*gomock.Call)",
	},
	"success-any-nay": {
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
	test.Map(t, callMatcherTestCases).
		Run(func(t test.Test, param MatcherParams) {
			// Given - send mock calls to unchecked test context.
			mocks := mock.NewMocks(test.New(t, false).Expect(test.Success))
			matcher := param.matcher(evalCall(param.base, mocks))

			// When
			matches := matcher.Matches(evalCall(param.match, mocks))

			// Then
			assert.Equal(t, param.expectMatches, matches)
			assert.Equal(t, param.expectString, matcher.String())
		})
}

type ReporterParams struct {
	setup  mock.SetupFunc
	misses func(test.Test, *mock.Mocks) mock.SetupFunc
	call   test.Func
	expect test.Expect
}

var reporterTestCases = map[string]ReporterParams{
	"log called": {
		setup: test.Log("log message"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Log("log message")
		},
		expect: test.Success,
	},
	"logf called": {
		setup: test.Logf("%s", "log message"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Logf("%s", "log message")
		},
		expect: test.Success,
	},
	"error called": {
		setup: test.Error("fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Error("fail")
		},
	},
	"errorf called": {
		setup: test.Errorf("%s", "fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Errorf("%s", "fail")
		},
	},
	"fatal called": {
		setup: test.Fatal("fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Fatal("fail")
		},
	},
	"fatalf called": {
		setup: test.Fatalf("%s", "fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Fatalf("%s", "fail")
		},
	},
	"fail called": {
		setup: test.Fail(),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Fail()
		},
	},
	"failnow called": {
		setup: test.FailNow(),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.FailNow()
		},
	},
	"panic called": {
		setup: test.Panic("fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			panic("fail")
		},
	},

	"log undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Log", CallerLog, "log"),
		call: func(t test.Test) {
			t.Log("log")
		},
	},
	"log undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Log", CallerLog, "log"),
		call: func(t test.Test) {
			t.Log("log")
			t.Log("log")
		},
	},
	"logf undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Logf", CallerLogf, "%s", "log"),
		call: func(t test.Test) {
			t.Logf("%s", "log")
		},
	},
	"logf undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Logf", CallerLogf, "%s", "log"),
		call: func(t test.Test) {
			t.Logf("%s", "log")
			t.Logf("%s", "log")
		},
	},
	"error undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Error", CallerError, "fail"),
		call: func(t test.Test) {
			t.Error("fail")
		},
	},
	"error undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Error", CallerError, "fail"),
		call: func(t test.Test) {
			t.Error("fail")
			t.Error("fail")
		},
	},
	"errorf undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Errorf", CallerErrorf, "%s", "fail"),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
	"errorf undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Errorf", CallerErrorf, "%s", "fail"),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
			t.Errorf("%s", "fail")
		},
	},
	"fatal undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fatal", CallerFatal, "fail"),
		call: func(t test.Test) {
			t.Fatal("fail")
		},
	},
	"fatal undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fatal", CallerFatal, "fail"),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing.
			t.Fatal("fail")
			t.Fatal("fail")
		},
	},
	"fatalf undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fatalf", CallerFatalf, "%s", "fail"),
		call: func(t test.Test) {
			t.Fatalf("%s", "fail")
		},
	},
	"fatalf undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fatalf", CallerFatalf, "%s", "fail"),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing
			t.Fatalf("%s", "fail")
			t.Fatalf("%s", "fail")
		},
	},
	"fail undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fail", CallerFail),
		call: func(t test.Test) {
			t.Fail()
		},
	},
	"fail undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Fail", CallerFail),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing
			t.Fail()
			t.Fail()
		},
	},
	"failnow undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"FailNow", CallerFailNow),
		call: func(t test.Test) {
			t.FailNow()
		},
	},
	"failnow undeclared twice": {
		misses: test.UnexpectedCall(test.NewValidator,
			"FailNow", CallerFailNow),
		call: func(t test.Test) {
			//revive:disable-next-line:unreachable-code // needed for testing
			t.FailNow()
			t.FailNow()
		},
	},
	"panic undeclared": {
		misses: test.UnexpectedCall(test.NewValidator,
			"Panic", CallerPanic, "fail"),
		call: func(test.Test) {
			panic("fail")
		},
	},

	// Only `Error`and `Errorf` can be consumed more than once, since `Fatal`,
	// `Fatalf`, `FailNow`, and panic will stop execution immediately. The
	// second call is effectively unreachable.
	"error consumed": {
		setup: test.Error("fail"),
		misses: test.ConsumedCall(test.NewValidator,
			"Error", CallerTestError, CallerReporterError, "fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Error("fail")
			t.Error("fail")
		},
	},
	"errorf consumed": {
		setup: test.Errorf("%s", "fail"),
		misses: test.ConsumedCall(test.NewValidator,
			"Errorf", CallerTestErrorf, CallerReporterErrorf, "%s", "fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			t.Errorf("%s", "fail")
			t.Errorf("%s", "fail")
		},
	},
	"fatal consumed": {
		setup: test.Fatal("fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			//revive:disable-next-line:unreachable-code // needed for testing.
			t.Fatal("fail")
			t.Fatal("fail")
		},
	},
	"fatalf consumed": {
		setup: test.Fatalf("%s", "fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			//revive:disable-next-line:unreachable-code // needed for testing.
			t.Fatalf("%s", "fail")
			t.Fatalf("%s", "fail")
		},
	},
	"fail consumed": {
		setup: test.Fail(),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			//revive:disable-next-line:unreachable-code // needed for testing.
			t.Fail()
			t.Fail()
		},
	},
	"failnow consumed": {
		setup: test.FailNow(),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			//revive:disable-next-line:unreachable-code // needed for testing.
			t.FailNow()
			t.FailNow()
		},
	},
	"panic consumed": {
		setup: test.Panic("fail"),
		call: func(t test.Test) {
			test.Cast[*test.Context](t).Expect(test.Success)
			panic("fail")
			//nolint:govet // needed for testing.
			panic("fail")
		},
	},

	// The mock setup is automatically creating a [test.Validator] requiring
	// a the test environment to expect a failure to get called. To satisfy
	// this, we need to create at least one failure.
	"errorf missing": {
		setup:  mock.Chain(test.Errorf("%s", "fail"), test.Errorf("%s", "fail")),
		misses: test.MissingCalls(test.Errorf("%s", "fail")),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
	"errorf missing two calls": {
		setup: mock.Chain(
			test.Errorf("%s", "fail"), test.Errorf("%s", "fail"),
			test.Errorf("%s", "fail-x"),
		),
		misses: test.MissingCalls(
			test.Errorf("%s", "fail"), test.Errorf("%s", "fail-x"),
		),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
	"fatalf missing": {
		setup:  mock.Chain(test.Errorf("%s", "fail"), test.Fatalf("%s", "fail")),
		misses: test.MissingCalls(test.Fatalf("%s", "fail")),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
	"fail missing": {
		setup:  mock.Chain(test.Errorf("%s", "fail"), test.Fail()),
		misses: test.MissingCalls(test.Fail()),
		call: func(t test.Test) {
			t.Fail()
		},
	},
	"failnow missing": {
		setup:  mock.Chain(test.Errorf("%s", "fail"), test.FailNow()),
		misses: test.MissingCalls(test.FailNow()),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
	"panic missing": {
		setup:  mock.Chain(test.Errorf("%s", "fail"), test.Panic("fail")),
		misses: test.MissingCalls(test.Panic("fail")),
		call: func(t test.Test) {
			t.Errorf("%s", "fail")
		},
	},
}

func TestReporter(t *testing.T) {
	test.Map(t, reporterTestCases).
		Run(func(t test.Test, param ReporterParams) {
			// Given
			mocks := mock.NewMocks(t)

			// When
			test.InRun(test.Success, func(tt test.Test) {
				// Given
				imocks := mock.NewMocks(tt)
				if param.misses != nil {
					mocks.Expect(param.misses(tt, imocks))
				}
				imocks.Expect(param.setup)

				// Connect the mock controller directly to the isolated parent
				// test environment to capture the mock controller failure.
				imocks.Ctrl.T = t

				// When
				param.call(tt)
			})(t)
		})
}
