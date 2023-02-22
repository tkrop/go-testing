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
	typeFileDflt  = &Type{File: DirDefault}
	typeFileTest  = &Type{File: pkgTest}
	typeFileIFace = &Type{File: filepath.Join(pkgTest, fileIFace)}
	typeFileTestX = &Type{File: filepath.Join(pkgTest, fileUnknown)}
	typeFileDirX  = &Type{File: dirUnknown}
	typeFileUpX   = &Type{File: filepath.Join(dirUp, fileUnknown)}
	typeFileTopX  = &Type{File: filepath.Join(dirTop, fileUnknown)}

	typeFileDirFileX = &Type{File: filepath.Join(dirUnknown, fileUnknown)}

	typePathDflt     = &Type{Path: DirDefault}
	typePathTest     = &Type{Path: pathTest}
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
		expect: targetMock,
	},

	"file default": {
		loader: loaderMock,
		target: typeFileDflt,
		expect: targetMock.With(typeFileDflt),
	},
	"file package replace": {
		loader: loaderMock,
		target: typeFileDflt.With(&Type{Package: pkgTest}),
		expect: targetMock.With(typeFileDflt),
	},
	"file package match": {
		loader: loaderMock,
		target: typeFileDflt.With(&Type{Package: pkgMockTest}),
		expect: targetMockTest.With(typeFileDflt),
	},
	"file package missing": {
		loader: loaderMock,
		target: typeFileDirFileX.With(&Type{Package: "unknown"}),
		expect: targetUnknown.With(typeFileDirFileX),
	},
	"file package missing test": {
		loader: loaderMock,
		target: typeFileDirFileX.With(&Type{Package: "unknown_test"}),
		expect: targetUnknown.With(typeFileDirFileX).
			With(&Type{Package: "unknown_test"}),
	},
	// TODO: add case with only <pkg>_test in package path.

	"file test": {
		loader: loaderMock,
		target: typeFileTest,
		expect: targetTest.With(typeFileTest),
	},
	"file iface": {
		loader: loaderMock,
		target: typeFileIFace,
		expect: targetTest.With(typeFileIFace),
	},
	"file child-missing": {
		loader: loaderMock,
		target: typeFileTestX,
		expect: targetTest.With(typeFileTestX),
	},
	"file missing": {
		loader: loaderMock,
		target: typeFileDirX,
		expect: targetInternal.With(typeFileDirX),
	},
	"file parent-missing": {
		loader: loaderMock,
		target: typeFileUpX,
		expect: targetInternal.With(typeFileUpX),
	},
	"file top-missing": {
		loader: loaderMock,
		target: typeFileTopX,
		expect: (&Type{}).With(typeFileTopX),
	},

	"path default": {
		loader: loaderMock,
		target: typePathDflt,
		expect: targetMock,
	},
	"path package replace": {
		loader: loaderMock,
		target: typePathDflt.With(&Type{Package: pkgTest}),
		expect: targetMock,
	},
	"path package match": {
		loader: loaderMock,
		target: typePathDflt.With(&Type{Package: pkgMockTest}),
		expect: targetMockTest,
	},

	"path test": {
		loader: loaderMock,
		target: typePathTest,
		expect: targetTest,
	},
	"path mock": {
		loader: loaderMock,
		target: typePathMock,
		expect: targetMock,
	},
	"path parent": {
		loader: loaderMock,
		target: typePathInternal,
		expect: targetInternal,
	},
	"path mock absolute": {
		loader: loaderMock,
		target: typePathMockAbs,
		expect: targetMock,
	},
	"path unknown absolute": {
		loader: loaderMock,
		target: typePathUnknownAbs,
		expect: targetUnknown,
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
	loader Loader
	source *Type
}

var testLoaderLoadParams = map[string]LoaderLoadParams{
	"file default": {
		loader: loaderMock,
		source: targetTest.With(typeFileDflt),
	},
}

func TestLoaderLoad(t *testing.T) {
	test.Map(t, testLoaderLoadParams).
		Run(func(t test.Test, param LoaderLoadParams) {
			// When
			param.loader.Load(param.source.Path)

			// Then
			// assert.Equal(t, param.expect, target)
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
