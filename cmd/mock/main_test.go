package main

import (
	"testing"

	"github.com/tkrop/go-testing/test"
)

var mainTestCases = map[string]test.MainParams{
	"no mocks": {
		Args:     []string{"mock"},
		Env:      []string{},
		ExitCode: 0,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, mainTestCases).Run(test.Main(main))
}
