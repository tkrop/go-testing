package test_test

import (
	"context"
	"fmt"
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
		Env:      []string{"panic=true"},
		Args:     []string{"panic"},
		ExitCode: 1,
	},
	"exit-1": {
		Env:      []string{"exit=1", "panic=false"},
		Args:     []string{"exit-1"},
		ExitCode: 1,
	},
	"exit-0": {
		Env:      []string{"exit=0", "panic=true", "panic=false"},
		Args:     []string{"exit-0"},
		ExitCode: 0,
	},
	"sleep": {
		Args:     []string{"sleep", "100ms"},
		ExitCode: 0,
	},
	"deadline": {
		Args: []string{"deadline", "1s"},
		Ctx: test.First(context.WithTimeout(context.Background(),
			time.Millisecond)),
		Error:    context.DeadlineExceeded,
		ExitCode: -1,
	},
	"interrupt": {
		Args: []string{"interrupt", "1s"},
		Ctx: test.First(context.WithTimeout(context.Background(),
			500*time.Millisecond)),
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

// ctx returns the current time formatted as RFC3339Nano truncated to 26
// characters to avoid excessive precision in test output.
func ctx() string {
	return fmt.Sprintf("%s [%s]", time.Now().Format(time.RFC3339Nano[0:26]), os.Args[0])
}

// main is a test main function to demonstrate the usage of `test.Main`.
func main() {
	// Check that environment variables are set correctly.
	fmt.Fprintf(os.Stderr, "%s var=%s\n", ctx(), os.Getenv("var"))
	if os.Getenv("panic") == "true" {
		fmt.Fprintf(os.Stderr, "%s var=%s\n", ctx(), os.Getenv("var"))
		panic("supposed to panic")
	}

	// Simulate some work.
	fmt.Fprintf(os.Stderr, "%s args=%v\n", ctx(), os.Args)
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "%s sleep=%s\n", ctx(), os.Args[1])
		dur, err := time.ParseDuration(os.Args[1])
		if err == nil {
			time.Sleep(dur)
		}
	}

	// Exit with given code.
	fmt.Fprintf(os.Stderr, "%s exit=%s\n", ctx(), os.Getenv("exit"))
	os.Exit(test.First(strconv.Atoi(os.Getenv("exit"))))
}
