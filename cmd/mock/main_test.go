package main

import (
	"testing"

	"github.com/tkrop/go-testing/test"
)

var testMainParams = map[string]test.MainParams{
	"no mocks": {
		Args:     []string{"mock"},
		Env:      []string{},
		ExitCode: 0,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).Run(test.TestMain(main))
}
