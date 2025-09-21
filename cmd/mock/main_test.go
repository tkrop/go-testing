package main

import (
	"testing"

	"github.com/tkrop/go-testing/test"
)

var testMainParams = map[string]test.MainParam{
	"no mocks": {
		Args:     []string{"mock"},
		Env:      []string{},
		ExitCode: 0,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).Run(test.Main(main))
}
