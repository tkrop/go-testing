package main

import (
	"fmt"
	"os"

	"github.com/tkrop/go-testing/internal/mock"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args[1:]...))
}

func run(stdout, stderr *os.File, args ...string) int {
	// Setup package loader and command line parser.
	loader := mock.NewLoader(mock.DefaultDir)
	target := mock.Type{File: "mock_all_test.go"}
	parser := mock.NewParser(loader, target)

	// Parse command line arguments to obtain mocks.
	mocks, errs := parser.Parse(args...)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintf(stderr, "%s\n", err)
		}
		return 1
	}

	// Generate mock files using the default template.
	temp, imports, err := mock.NewTemplate()
	if err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		return 2
	}

	failure := false
	files := mock.NewFiles(mocks, imports...)
	for _, file := range files {
		if err := file.Open(stdout); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}

		if err := file.Write(temp); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}

		if err := file.Close(); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}
	}

	if failure {
		return 3
	}
	return 0
}
