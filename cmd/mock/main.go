package main

import (
	"fmt"
	"os"

	"github.com/tkrop/go-testing/internal/mock"
)

func main() {
	wd, err := os.Getwd()
	fmt.Fprintf(os.Stdout, "working directory: %s (%v)\n", wd, err)
	fmt.Fprintf(os.Stdout, "command-line-args: %v\n", os.Args)
	gen := mock.NewGenerator(mock.DirDefault, mock.TargetDefault)
	os.Exit(gen.Generate(os.Stdout, os.Stdout, os.Args[1:]...))
}
