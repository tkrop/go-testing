package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/stretchr/testify/require"
)

// Must is a convenience method returning the value of the first argument and
// that panics on any error in the second argument using the provided error.
// The method allows to write concise test setup code.
func Must[T any](arg T, err error) T {
	if err != nil {
		panic(err)
	}
	return arg
}

// Cast is a convenience function to cast the given argument to the specified
// type or panic if the cast fails. The method allows to write concise test
// setup code granting meaningful type checks.
func Cast[T any](arg any) T {
	val, ok := arg.(T)
	if !ok {
		panic(fmt.Sprintf("cast failed [%T]: %v", val, arg))
	}
	return val
}

// First is a convenience function to return the first argument and ignore all
// others arguments. The method allows to write concise test setup code.
func First[T any](arg T, _ ...any) T { return arg }

// TODO: consider following convenience methods:
//
// // Check is a convenience method that returns the second argument and swallows
// // the first used to focus a test on the second.
// func Check[T any](swallowed any, check T) T {
// 	return check
// }

// // Ok is a convenience method to check whether the second boolean argument is
// // `true` while returning the first argument. If the boolean argument is
// // `false`, the method panics.
// func Ok[T any](result T, ok bool) T {
// 	if !ok {
// 		panic("bool not okay")
// 	}
// 	return result
// }

// MainParam provides the test parameters for testing a `main`-method.
type MainParam struct {
	// Ctx is the context to control the test process, e.g. for deadlines and
	// cancellations. If not set, a background context is used with unlimited
	// duration.
	Ctx context.Context //nolint:containedctx // only in test parameters.

	// Args are the command line arguments to provide to the test process. The
	// first argument is usually the program name, typically `os.Args[0]`. If
	// not provided, the `main`-method is called with no arguments at all.
	Args []string

	// Env are the additional environment variables that are provided to the
	// spawned test process. The existing environment variables are inherited
	// and an additional `GO_TESTING_TEST` variable is set to select the test
	// case.
	Env []string

	// ExitCode is the expected exit code when running the test process. If
	// the exit code is not `0`, the error must be of type `*exec.ExitError`.
	ExitCode int

	// Error is the expected error when running the test process. This is only
	// used in edge cases when the test error is not of type `*exec.ExitError`.
	Error error
}

// GoTestingRunVar is the environment variable used to signal the new process
// to execute the `main` method instead of spawning a new test process.
const GoTestingRunVar = "GO_TESTING_RUN"

// Main creates a test function that runs the given `main`-method in a separate
// test process to protect the test execution from `os.Exit` calls while allowing
// to capture and check the exit code against the expectation. The following
// example demonstrates how to use this method to test a `main`-method:
//
//	 mainTestCases := map[string]test.MainParam{
//		 "with args": {
//			 Args: []string{"mock", "arg1", "arg2"},
//			 Env:  []string{"VAR=value"},
//			 ExitCode: 0,
//		 },
//	 }
//
//	 func Main(t *testing.T) {
//		 test.Map(t, mainTestCases).Run(test.Main(main))
//	 }
//
// If the test process is expected to run longer than the default test timeout,
// a context with timeout can be provided to interrupt the test process in time.
// This e.g. can be done as follows using `test.First` to ignore the cancelFunc:
//
//	Ctx: test.First(context.WithTimeout(context.Background(), time.Second))
func Main(main func()) func(t Test, param MainParam) {
	return func(t Test, param MainParam) {
		// Switch to execute main function in test process.
		if name := os.Getenv(GoTestingRunVar); name != "" {
			// Ensure only expected test is running.
			if name == t.Name() {
				os.Args = param.Args
				main()
				require.Fail(t, "os-exit not called")
			}
			// Skip unexpected tests.
			return
		}

		// Prepare environment for the test process.
		ctx := context.Background()
		if param.Ctx != nil {
			ctx = param.Ctx
		}

		// #nosec G204 -- secured by calling only the test instance.
		cmd := exec.CommandContext(ctx, os.Args[0],
			"-test.run="+t.(*Context).t.Name())

		// No stdout to allow propagation of coverage results.
		cmd.Stdin, cmd.Stderr = os.Stdin, os.Stderr
		cmd.Env = append(os.Environ(), append(param.Env,
			GoTestingRunVar+"="+t.Name())...)

		if err := cmd.Run(); err != nil || param.ExitCode != 0 {
			errExit := &exec.ExitError{}
			if errors.As(err, &errExit) {
				require.Equal(t, param.ExitCode, errExit.ExitCode())
			} else if err != nil {
				require.Equal(t, param.Error, err)
			}
		}
		require.Equal(t, param.ExitCode, cmd.ProcessState.ExitCode())
	}
}
