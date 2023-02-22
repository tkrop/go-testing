package mock

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

var ErrFileOpening = errors.New("file opening")

func NewErrFileOpening(path string, err error) error {
	return fmt.Errorf("%w [name: %s]: %w",
		ErrFileOpening, path, err)
}

var ErrNotFound = errors.New("not found")

func NewErrNotFound(source *Type, name string) error {
	return fmt.Errorf("%w [path: %s (%s), name: %s]",
		ErrNotFound, source.Path, source.File, name)
}

var ErrNoNameType = errors.New("no name type")

func NewErrNoNameType(source *Type, name string) error {
	return fmt.Errorf("%w [path: %s (%s), name: %s]",
		ErrNoNameType, source.Path, source.File, name)
}

var ErrNoIFace = errors.New("no interface")

func NewErrNoIFace(source *Type, name string) error {
	return fmt.Errorf("%w [path: %s (%s), name: %s]",
		ErrNoIFace, source.Path, source.File, name)
}

var ErrMatcherInvalid = errors.New("matcher invalid")

func NewErrMatcherInvalid(source *Type, err error) error {
	return fmt.Errorf("%w [file: %s, name: %s]: %w",
		ErrMatcherInvalid, source.File, source.Name, err)
}

var ErrLoading = errors.New("loading")

func NewErrLoading(path string, err error) error {
	return fmt.Errorf("%w [path: %s] => %w",
		ErrLoading, path, err)
}

var ErrArgInvalid = errors.New("argument invalid")

func NewErrArgInvalid(pos int, arg string) error {
	return fmt.Errorf("%w [pos: %d, arg: %s]",
		ErrArgInvalid, pos, arg)
}

func NewErrArgNotFound(pos int, arg string) error {
	return fmt.Errorf("%w [pos: %d, arg: %s]: %v",
		ErrArgInvalid, pos, arg, ErrNotFound)
}

var ErrArgFailure = errors.New("argument failure")

func NewErrArgFailure(pos int, arg string, err error) error {
	return fmt.Errorf("%w [pos: %d, arg: %s] => %v",
		ErrArgFailure, pos, arg, err)
}

var ErrAliasConflict = errors.New("alias conflict")

func NewErrAliasConflict(imprt *Import, path string) error {
	return fmt.Errorf("%w [alias: %s, path: %s <=> %s]",
		ErrAliasConflict, imprt.Alias, imprt.Path, path)
}

var ErrIllegalImport = errors.New("illegal import")

func NewErrIllegalImport(imprt *Import) error {
	return fmt.Errorf("%w [alias: %s, path: %s]",
		ErrIllegalImport, imprt.Alias, imprt.Path)
}

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
