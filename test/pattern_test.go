package test_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type MustParam struct {
	setup  mock.SetupFunc
	arg    any
	err    error
	expect any
}

var testMustParams = map[string]MustParam{
	"nil": {},
	"string": {
		arg:    "value",
		expect: "value",
	},
	"integer": {
		arg:    1,
		expect: 1,
	},
	"error": {
		setup:  test.Panic(assert.AnError),
		err:    assert.AnError,
		expect: nil,
	},
}

func TestMust(t *testing.T) {
	test.Map(t, testMustParams).Run(func(t test.Test, param MustParam) {
		// Given
		mock.NewMocks(t).Expect(param.setup)

		// When
		result := test.Must(param.arg, param.err)

		// Then
		assert.Equal(t, param.expect, result)
	})
}

var testMainParams = map[string]test.MainParam{
	"panic": {
		Env:      []string{"exit=2", "TEST=other"},
		ExitCode: 1,
	},
	"exit-0": {
		Env:      []string{"exit=0", "var=1", "var=2"},
		ExitCode: 0,
	},
	"exit-1": {
		Env:      []string{"exit=1", "var=2"},
		ExitCode: 1,
	},
	"sleep": {
		Args:     []string{"100ms"},
		Env:      []string{"var=2"},
		ExitCode: 0,
	},
	"deadline": {
		Ctx: test.First(context.WithTimeout(context.Background(),
			time.Millisecond)),
		Args:     []string{"1s"},
		Env:      []string{"var=2"},
		Error:    context.DeadlineExceeded,
		ExitCode: -1,
	},
	"interrupt": {
		Ctx: test.First(context.WithTimeout(context.Background(),
			500*time.Millisecond)),
		Args:     []string{"1s"},
		Env:      []string{"var=2"},
		ExitCode: -1,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).Run(test.Main(main))
}

func TestMainUnexpected(t *testing.T) {
	t.Setenv(test.GoTestingRunVar, "other")
	test.Param(t, test.MainParam{}).RunSeq(test.Main(main))
}

func main() {
	// Get the expected exit code from the environment.
	exit, _ := strconv.Atoi(os.Getenv("exit"))

	// Check that environment variables are set correctly.
	if os.Getenv("var") != "2" {
		panic("env var not set")
	}

	// Simulate some work.
	if len(os.Args) > 0 {
		dur, err := time.ParseDuration(os.Args[0])
		if err == nil {
			time.Sleep(dur)
		}
	}

	os.Exit(exit)
}
