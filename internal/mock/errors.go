package mock

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

var ErrPackageParsing = errors.New("package parsing")

func NewErrPackageParsing(path string, pkgs []*packages.Package) error {
	errs := []packages.Error{}
	for _, pkg := range pkgs {
		if len(pkg.Errors) != 0 {
			errs = append(errs, pkg.Errors...)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("%w [path: %s] => %v",
			ErrPackageParsing, path, errs)
	}
	return nil
}

var ErrNotFound = errors.New("not found")

func NewErrNotFound(path, name string) error {
	return fmt.Errorf("%w [path: %s, name: %s]",
		ErrNotFound, path, name)
}

var ErrNoNameType = errors.New("no name type")

func NewErrNoNameType(path, name string) error {
	return fmt.Errorf("%w [path: %s, name: %s]",
		ErrNoNameType, path, name)
}

var ErrNoIFace = errors.New("no interface")

func NewErrNoIFace(path, name string) error {
	return fmt.Errorf("%w [path: %s, name: %s]",
		ErrNoIFace, path, name)
}

var ErrLoading = errors.New("loading")

func NewErrLoading(path string, err error) error {
	return fmt.Errorf("%w [path: %s] => %v",
		ErrLoading, path, err)
}

var ErrInvalidArg = errors.New("argument invalid")

func NewErrArgInvalid(pos int, arg string) error {
	return fmt.Errorf("%w [pos: %d, arg: %s]",
		ErrInvalidArg, pos, arg)
}

var ErrArgFailure = errors.New("argument failure")

func NewErrArgFailure(pos int, arg string, err error) error {
	return fmt.Errorf("%w [pos: %d, arg: %s] => %v",
		ErrArgFailure, pos, arg, err)
}

func NewErrAliasConflict(path, alias string) error {
	return fmt.Errorf("alias conflict [path: %s, alias: %s]",
		path, alias)
}
