package test_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/tkrop/go-testing/test"
)

var testMainParams = map[string]test.MainParams{
	"exit-0": {},
	"exit-1": {
		Env:      []string{"exit=1"},
		ExitCode: 1,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).Run(test.TestMain(main))
}

func TestMainUnexpected(t *testing.T) {
	t.Setenv("TEST", "other")
	test.Map(t, testMainParams).RunSeq(test.TestMain(main))
}

func main() {
	exit, _ := strconv.Atoi(os.Getenv("exit"))
	os.Exit(exit)
}
