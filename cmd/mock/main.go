package main

import (
	"os"

	"github.com/tkrop/go-testing/internal/mock"
)

func main() {
	gen := mock.NewGenerator(mock.DirDefault, mock.TargetDefault)
	os.Exit(gen.Generate(os.Stdout, os.Stderr, os.Args[1:]...))
}
