package test

import (
	"errors"
	"os"
	"os/exec"

	"github.com/stretchr/testify/require"
)

// MainParams provides the test parameters for testing a `main`-method.
type MainParams struct {
	Args     []string
	Env      []string
	ExitCode int
}

// Main creates a test function that runs the given `main`-method in a
// separate test process to allow capturing the exit code and checking it
// against the expectation. This can be applied as follows:
//
//	 testMainParams := map[string]test.MainParams{
//		 "no mocks": {
//			 Args:     []string{"mock"},
//			 Env:      []string{"VAR=value"},
//			 ExitCode: 0,
//		 },
//	 }
//
//	 func Main(t *testing.T) {
//		 test.Map(t, testMainParams).Run(test.Main(main, "VAR=value"...))
//	 }
//
// The test method is spawning a new test using the `TEST` environment variable
// to select the expected parameter set containing the command line arguments
// (`Args`) to execute in the spawned process. The test instance is also called
// setting the given additional environment variables (`Env`) to allow
// modification of the test environment.
func Main(main func(), env ...string) func(t Test, param MainParams) {
	return func(t Test, param MainParams) {
		// Switch to execute main function in test process.
		if name := os.Getenv("TEST"); name != "" {
			// Ensure only expected test is running.
			if name == t.Name() {
				os.Args = param.Args
				main()
				require.Fail(t, "os-exit not called")
			}
			// Skip unexpected tests.
			return
		}

		// Call the main function in a separate process to prevent capture
		// regular process exit behavior.
		// #nosec G204 -- secured by calling only the test instance.
		cmd := exec.Command(os.Args[0], "-test.run="+t.(*Context).t.Name())
		cmd.Env = append(append(os.Environ(), "TEST="+t.Name()),
			append(env, param.Env...)...)
		if err := cmd.Run(); err != nil || param.ExitCode != 0 {
			errExit := &exec.ExitError{}
			if errors.As(err, &errExit) || err != nil {
				require.Equal(t, param.ExitCode, errExit.ExitCode())
			} else {
				// #no-cover: impossible to reach this code.
				require.Fail(t, "unexpected error", err)
			}
		}
	}
}
