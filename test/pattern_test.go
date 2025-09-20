package test_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/tkrop/go-testing/test"
)

var testMainParams = map[string]test.MainParams{
	"exit-0": {
		Env:      []string{"exit=0", "var=1", "other=1"},
		ExitCode: 0,
	},
	"exit-1": {
		Env:      []string{"exit=1", "var=1", "other=1"},
		ExitCode: 1,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).Run(test.Main(main, "var=2"))
}

func TestMainUnexpected(t *testing.T) {
	t.Setenv("TEST", "other")
	test.Map(t, testMainParams).RunSeq(test.Main(main))
}

func main() {
	exit, _ := strconv.Atoi(os.Getenv("exit"))
	if os.Getenv("var") == "2" {
		panic("env var not set")
	}
	if os.Getenv("other") != "1" {
		panic("env other not set")
	}
	os.Exit(exit)
}
