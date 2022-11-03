package test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
)

var testRunParams = map[string]struct {
	expect Expect
	test   func(*TestingT)
}{
	"run success": {
		test: InRun(ExpectSuccess,
			func(t *TestingT) {}),
		expect: ExpectSuccess,
	},

	"run success with errorf": {
		test: InRun(ExpectSuccess,
			func(t *TestingT) { t.Errorf("fail") }),
		expect: ExpectFailure,
	},

	"run success with fatalf": {
		test: InRun(ExpectSuccess,
			func(t *TestingT) { t.Fatalf("fail") }),
		expect: ExpectFailure,
	},

	"run success with failnow": {
		test: InRun(ExpectSuccess,
			func(t *TestingT) { t.FailNow() }),
		expect: ExpectFailure,
	},

	"run failure with errorf": {
		test: InRun(ExpectFailure,
			func(t *TestingT) { t.Errorf("fail") }),
		expect: ExpectSuccess,
	},

	"run failure with fatalf": {
		test: InRun(ExpectFailure,
			func(t *TestingT) { t.Fatalf("fail") }),
		expect: ExpectSuccess,
	},

	"run failure with failnow": {
		test: InRun(ExpectFailure,
			func(t *TestingT) { t.FailNow() }),
		expect: ExpectSuccess,
	},

	"in success": {
		test:   InSuccess(func(t *TestingT) {}),
		expect: ExpectSuccess,
	},

	"in success with errorf": {
		test: InSuccess(
			func(t *TestingT) { t.Errorf("fail") }),
		expect: ExpectFailure,
	},

	"in success with fatalf": {
		test: InSuccess(
			func(t *TestingT) { t.Fatalf("fail") }),
		expect: ExpectFailure,
	},

	"in success with failnow": {
		test: InSuccess(
			func(t *TestingT) { t.FailNow() }),
		expect: ExpectFailure,
	},

	"in failure with errorf": {
		test: InFailure(
			func(t *TestingT) { t.Errorf("fail") }),
		expect: ExpectSuccess,
	},

	"in failure with fatalf": {
		test: InFailure(
			func(t *TestingT) { t.Fatalf("fail") }),
		expect: ExpectSuccess,
	},

	"in failure with failnow": {
		test: InFailure(
			func(t *TestingT) { t.FailNow() }),
		expect: ExpectSuccess,
	},
}

func TestRun(t *testing.T) {
	for message, param := range testRunParams {
		t.Run(message, Run(param.expect, func(t *TestingT) {
			require.NotEmpty(t, message)

			// Given
			wg := mock.NewMock(t).WaitGroup()
			t.WaitGroup(wg)
			go func() { wg.Add(1); wg.Wait() }()

			// When
			param.test(t)

			// Then
			wg.Add(math.MinInt)
		}))
	}
}

func TestOther(t *testing.T) {
	for message, param := range testRunParams {
		switch param.expect {
		case ExpectSuccess:
			t.Run(message, Success(func(t *TestingT) {
				require.NotEmpty(t, message)
				param.test(t)
			}))

		case ExpectFailure:
			t.Run(message, Failure(func(t *TestingT) {
				require.NotEmpty(t, message)
				param.test(t)
			}))
		}
	}
}
