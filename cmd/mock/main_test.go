package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/test"
)

type MainParams struct {
	args           []string
	expectExitCode int
}

var testMainParams = map[string]MainParams{
	"no mock to generate": {
		args:           []string{"mock"},
		expectExitCode: 0,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).
		Run(func(t test.Test, param MainParams) {
			// Switch to execute main function in test process.
			if name := os.Getenv("TEST"); name != "" {
				// Ensure only expected test is running.
				if name == t.Name() {
					os.Args = param.args
					main()
					assert.Fail(t, "os-exit not called")
				}
				// Skip other test.
				return
			}

			// Call the main function in a separate process to prevent capture
			// regular process exit behavior.
			cmd := exec.Command(os.Args[0], "-test.run=TestMain")
			cmd.Env = append(os.Environ(), "TEST="+t.Name())
			if err := cmd.Run(); err != nil || param.expectExitCode != 0 {
				errExit := &exec.ExitError{}
				if errors.As(err, &errExit) {
					assert.Equal(t, param.expectExitCode, errExit.ExitCode())
				} else {
					assert.Fail(t, "unexpected error", err)
				}
			}
		})
}
