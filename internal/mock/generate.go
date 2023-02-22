package mock

import (
	"fmt"
	"os"
)

const (
	// Default mock file name.
	MockFileDefault = "mock_all_test.go"
)

// TargetDefault provides the default target setup for the parser.
var TargetDefault = &Type{File: MockFileDefault}

// Generator for mocks.
type Generator struct {
	parser   *Parser
	imports  []*Import
	template Template
}

// NewGenerator creates a new mock generator with given default context
// directory and default target setup.
func NewGenerator(dir string, target *Type) *Generator {
	tempplate, imports, _ := NewTemplate()
	return &Generator{
		parser:   NewParser(NewLoader(dir), target),
		template: tempplate,
		imports:  imports,
	}
}

// Generate creates the mock files defined by the command line arguments.
func (gen *Generator) Generate(stdout, stderr *os.File, args ...string) int {
	// Parse command line arguments to obtain mocks.
	mocks, errs := gen.parser.Parse(args...)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintf(stderr, "%s\n", err)
		}
		return 1
	}

	// Generate mock files using the default template.
	failure := false
	files := NewFiles(mocks, gen.imports...)
	for _, file := range files {
		if err := file.Open(stdout); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}

		if err := file.Write(gen.template); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}

		if err := file.Close(); err != nil {
			fmt.Fprintf(stderr, "%s\n", err)
			failure = true
		}
	}

	if failure {
		return 2
	}
	return 0
}
