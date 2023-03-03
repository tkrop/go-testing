package mock_test

import (
	"fmt"
	"path/filepath"
	"regexp/syntax"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

var (
	typePathDflt     = &Type{Path: DirDefault}
	typePathTest     = &Type{Path: pathTest}
	typePathTestX    = &Type{Path: pathTest + "x"}
	typePathMock     = &Type{Path: pathMock}
	typePathInternal = &Type{Path: pathInternal}
	typePathMockAbs  = &Type{Path: func() string {
		path, _ := filepath.Abs(DirDefault)
		return path
	}()}
	typePathUnknownAbs = &Type{Path: func() string {
		path, _ := filepath.Abs(dirUnknown)
		return path
	}()}
	typePathTopAbs = &Type{Path: func() string {
		path, _ := filepath.Abs(dirTop)
		return path
	}()}

	typePkgUnknown      = &Type{Package: pkgUnknown}
	typePkgUnknownTest  = &Type{Package: pkgUnknownTest}
	typePkgInternal     = &Type{Package: pkgInternal}
	typePkgInternalTest = &Type{Package: pkgInternalTest}

	typeFileDflt   = &Type{File: DirDefault}
	typeFileTest   = &Type{File: pkgTest}
	typeFileIFace  = &Type{File: filepath.Join(pkgTest, fileIFace)}
	typeFileIFaceX = &Type{File: filepath.Join(pkgTest+"x", fileIFace)}
	typeFileTestX  = &Type{File: filepath.Join(pkgTest, fileUnknown)}
	typeFileDirX   = &Type{File: dirUnknown}
	typeFileUpX    = &Type{File: filepath.Join(dirUp, fileUnknown)}
	typeFileTopX   = &Type{File: filepath.Join(dirTop, fileUnknown)}

	typeFileDirFileX = &Type{File: filepath.Join(dirUnknown, fileUnknown)}
)

type LoaderSearchParams struct {
	loader Loader
	target *Type
	expect *Type
}

var testLoaderSearchParams = map[string]LoaderSearchParams{
	"fail": {
		loader: loaderFail,
		target: typePathDflt,
		expect: typePathDflt,
	},
	"empty": {
		loader: loaderMock,
		expect: targetMockTest,
	},

	"package default": {
		loader: loaderMock,
		target: typeFileDflt,
		expect: targetMockTest.With(typeFileDflt),
	},
	"package replace": {
		loader: loaderMock,
		target: typeFileDflt.With(&Type{Package: pkgTest}),
		expect: targetMockTest.With(typeFileDflt),
	},
	"package match": {
		loader: loaderMock,
		target: typeFileIFaceX.With(typePkgTest),
		expect: targetTest.With(typePathTestX).With(typeFileIFaceX),
	},
	"package match test": {
		loader: loaderMock,
		target: typeFileIFaceX.With(typePkgTestTest),
		expect: targetTestTest.With(typePathTestX).With(typeFileIFaceX),
	},
	"package unknown": {
		loader: loaderMock,
		target: typeFileDirFileX.With(typePkgUnknown),
		expect: targetUnknown.With(typeFileDirFileX),
	},
	"package unknown test": {
		loader: loaderMock,
		target: typeFileDirFileX.With(typePkgUnknownTest),
		expect: targetUnknownTest.With(typeFileDirFileX),
	},
	"package internal": {
		loader: loaderMock,
		target: typeFileUpX.With(typePkgInternal),
		expect: targetInternal.With(typeFileUpX),
	},
	"package internal test": {
		loader: loaderMock,
		target: typeFileUpX.With(typePkgInternalTest),
		expect: targetInternalTest.With(typeFileUpX),
	},

	"file test": {
		loader: loaderMock,
		target: typeFileTest,
		expect: targetTestTest.With(typeFileTest),
	},
	"file iface": {
		loader: loaderMock,
		target: typeFileIFace,
		expect: targetTestTest.With(typeFileIFace),
	},
	"file child-missing": {
		loader: loaderMock,
		target: typeFileTestX,
		expect: targetTestTest.With(typeFileTestX),
	},
	"file missing": {
		loader: loaderMock,
		target: typeFileDirX,
		expect: targetInternalTest.With(typeFileDirX),
	},
	"file parent-missing": {
		loader: loaderMock,
		target: typeFileUpX,
		expect: targetInternalTest.With(typeFileUpX),
	},
	"file top-missing": {
		loader: loaderMock,
		target: typeFileTopX,
		expect: (&Type{}).With(typeFileTopX),
	},

	"path default": {
		loader: loaderMock,
		target: typePathDflt,
		expect: targetMockTest,
	},
	"path package replace": {
		loader: loaderMock,
		target: typePathDflt.With(&Type{Package: pkgTest}),
		expect: targetMockTest,
	},
	"path package match": {
		loader: loaderMock,
		target: typePathDflt.With(&Type{Package: pkgMockTest}),
		expect: targetMockTest,
	},

	"path test": {
		loader: loaderMock,
		target: typePathTest,
		expect: targetTestTest,
	},
	"path mock": {
		loader: loaderMock,
		target: typePathMock,
		expect: targetMockTest,
	},
	"path parent": {
		loader: loaderMock,
		target: typePathInternal,
		expect: targetInternalTest,
	},
	"path mock absolute": {
		loader: loaderMock,
		target: typePathMockAbs,
		expect: targetMockTest,
	},
	"path unknown absolute": {
		loader: loaderMock,
		target: typePathUnknownAbs,
		expect: targetUnknownTest,
	},
	"path top absolute": {
		loader: loaderMock,
		target: typePathTopAbs,
		expect: typePathTopAbs,
	},
}

func TestLoaderSearch(t *testing.T) {
	test.Map(t, testLoaderSearchParams).
		Run(func(t test.Test, param LoaderSearchParams) {
			// Given
			target := param.target.Copy()

			// When
			target.Update(param.loader)

			// Then
			assert.Equal(t, param.expect, target)
		})
}

type LoaderLoadParams struct {
	loader      Loader
	source      *Type
	expectError error
	expectLen   int
}

var testLoaderLoadParams = map[string]LoaderLoadParams{
	"file loading": {
		loader:    loaderMock,
		source:    targetTest.With(typeFileDflt),
		expectLen: 1,
	},
	"failure loading": {
		loader: loaderFail,
		source: targetTest.With(&Type{
			File: filepath.Join(dirUp, dirMock, dirTest, fileIFace),
		}),
		expectError: NewErrLoading(pathTest, fmt.Errorf(
			"err: chdir %s: no such file or directory: stderr: ",
			dirUnknown)),
		expectLen: 0,
	},
}

func TestLoaderLoad(t *testing.T) {
	test.Map(t, testLoaderLoadParams).
		Run(func(t test.Test, param LoaderLoadParams) {
			// When
			pkgs, err := param.loader.Load(param.source.Path).Get()

			// Then
			assert.Equal(t, param.expectError, err)
			assert.Len(t, pkgs, param.expectLen)
		})
}

type LoaderIFacesParams struct {
	loader       Loader
	source       *Type
	expectIFaces []*IFace
	expectError  error
}

var testLoaderIFacesParams = map[string]LoaderIFacesParams{
	"file default": {
		loader: loaderMock,
		source: targetTest.With(typeFileDflt),
		expectIFaces: []*IFace{{
			Source:  sourceIFaceAny,
			Methods: methodsLoadIFace,
		}},
	},
	"file iface": {
		loader: loaderMock,
		source: targetTest.With(typeFileIFace),
		expectIFaces: []*IFace{{
			Source:  sourceIFaceAny,
			Methods: methodsLoadIFace,
		}},
	},
	"file unknown": {
		loader: loaderMock,
		source: targetTest.With(typeFileTestX),
		expectError: NewErrNotFound(
			targetTest.With(typeFileTestX), ""),
	},
	"file unknown with name": {
		loader: loaderMock,
		source: targetTest.With(typeFileTestX).With(nameIFace),
		expectError: NewErrNotFound(
			targetTest.With(typeFileTestX), iface),
	},

	"name iface": {
		loader: loaderMock,
		source: targetTest.With(nameIFace),
		expectIFaces: []*IFace{{
			Source:  sourceIFaceAny,
			Methods: methodsLoadIFace,
		}},
	},
	"name pattern": {
		loader: loaderMock,
		source: targetTest.With(&Type{Name: MatchPatternDefault}),
		expectIFaces: []*IFace{{
			Source:  sourceIFaceAny,
			Methods: methodsLoadIFace,
		}},
	},
	"name pattern failure": {
		loader: loaderMock,
		source: targetTest.With(&Type{Name: "**"}),
		expectError: NewErrMatcherInvalid(
			targetTest.With(&Type{Name: "**"}),
			&syntax.Error{
				Code: "missing argument to repetition operator",
				Expr: "*",
			}),
	},
	"name struct": {
		loader: loaderMock,
		source: targetTest.With(&Type{Name: "Struct"}),
		expectError: NewErrNoIFace(targetTest.
			With(&Type{Name: "Struct"}), "Struct"),
	},
	"name missing": {
		loader: loaderMock,
		source: targetTest.With(&Type{Name: "Missing"}),
		expectError: NewErrNotFound(targetTest.
			With(&Type{Name: "Missing"}), "Missing"),
	},

	"failure loading": {
		loader: loaderFail,
		source: targetTest.With(&Type{
			File: filepath.Join(dirUp, dirMock, dirTest, fileIFace),
		}),
		expectError: NewErrLoading(pathTest, fmt.Errorf(
			"err: chdir %s: no such file or directory: stderr: ",
			dirUnknown)),
	},
}

func TestLoaderIFaces(t *testing.T) {
	test.Map(t, testLoaderIFacesParams).
		Run(func(t test.Test, param LoaderIFacesParams) {
			// Given
			source := param.source.Copy()

			// When
			ifaces, err := param.loader.IFaces(source)

			// Then
			assert.Equal(t, param.expectError, err)
			assert.Equal(t, param.expectIFaces, ifaces)
		})
}
