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
	typeFileMock    = &Type{File: fileMock}
	typeFileTemp    = &Type{File: filepath.Join(dirUp, fileTemplate)}
	typePkgTestTest = &Type{Package: pkgTestTest}
)

type ParseParams struct {
	loader      Loader
	target      *Type
	args        []string
	expectMocks []*Mock
	expectError []error
}

func testParse(t test.Test, param ParseParams) {
	// Given
	parser := NewParser(param.loader, param.target)

	// When
	mocks, errs := parser.Parse(param.args...)

	// Then
	assert.Equal(t, param.expectError, errs)
	assert.Equal(t, param.expectMocks, mocks)
}

var testParseParams = map[string]ParseParams{
	"no argument": {
		loader: loaderTest,
		args:   []string{},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"invalid argument": {
		loader:      loaderTest,
		args:        []string{"--unknown=any"},
		expectError: []error{NewErrArgInvalid(0, "--unknown=any")},
	},

	"default file": {
		loader: loaderTest,
		args:   []string{},
		target: typeFileMock,
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"default path": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Path: pathTesting},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestingIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"default package": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Package: pkgTestTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typePkgTestTest),
			Methods: methodsLoadIFace,
		}},
	},
	"default interface (ignore)": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Name: ifaceMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"source package explicit": {
		loader: loaderTest,
		args:   []string{"--source-pkg=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source package derived": {
		loader: loaderTest,
		args:   []string{"--source=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"source path explicit": {
		loader: loaderTest,
		args:   []string{"--source-path=" + pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path derived": {
		loader: loaderTest,
		args:   []string{"--source=" + pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path guessed": {
		loader: loaderMock,
		args:   []string{pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path default": {
		loader: loaderTest,
		args:   []string{DirDefault},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path invalid": {
		loader: loaderMock,
		args:   []string{pathUnknown},
		expectError: []error{
			NewErrArgNotFound(0, pathUnknown),
		},
	},

	"source file explicit": {
		loader: loaderMock,
		args:   []string{"--source-file=" + dirTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file derived": {
		loader: loaderMock,
		args:   []string{"--source=" + dirTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file derived missing": {
		loader: loaderMock,
		args: []string{
			"--source=" + fileUnknown,
		},
		expectError: []error{
			NewErrArgNotFound(0, fileUnknown),
		},
		expectMocks: nil,
	},
	"source file guessed": {
		loader: loaderMock,
		args:   []string{dirTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file guessed invalid": {
		loader: loaderMock,
		args:   []string{fileUnknown},
		expectError: []error{
			NewErrArgNotFound(0, fileUnknown),
		},
	},

	"source directory explicit": {
		loader: loaderMock,
		args:   []string{"--source-file=" + dirTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory derived": {
		loader: loaderMock,
		args:   []string{"--source=" + dirTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory guessed": {
		loader: loaderMock,
		args:   []string{dirTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory invalid": {
		loader: loaderMock,
		args:   []string{dirUnknown},
		expectError: []error{
			NewErrArgNotFound(0, dirUnknown),
		},
	},

	"target package explicit": {
		loader: loaderTest,
		args:   []string{"--target-pkg=" + pkgTestTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typePkgTestTest),
			Methods: methodsLoadIFace,
		}},
	},
	"target package derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pkgTestTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typePkgTestTest),
			Methods: methodsLoadIFace,
		}},
	},
	"target package guessed": {
		loader: loaderTest,
		args:   []string{pkgTestTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typePkgTestTest),
			Methods: methodsLoadIFace,
		}},
	},

	"target path explicit": {
		loader: loaderTest,
		args:   []string{"--target-path=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"target path derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"target file explicit": {
		loader: loaderTest,
		args:   []string{"--target-file=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file derived": {
		loader: loaderTest,
		args:   []string{"--target=" + typeFileTemp.File},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(typeFileTemp),
			Methods: methodsLoadIFace,
		}},
	},
	"target file derived missing": {
		loader: loaderTest,
		args:   []string{"--target=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file guessed": {
		loader: loaderTest,
		args:   []string{fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file with path guessed": {
		loader: loaderTest,
		args:   []string{filepath.Join(dirUp, fileMock)},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetMockIFace.With(&Type{
				File: filepath.Join(dirUp, fileMock),
			}),
			Methods: methodsLoadIFace,
		}},
	},

	"target file with package guessed": {
		loader: loaderTest,
		args:   []string{fileMock, pkgTestTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},

	"iface explicit": {
		loader: loaderTest,
		args:   []string{"--iface=" + iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed": {
		loader: loaderTest,
		args:   []string{iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed with mock": {
		loader: loaderTest,
		args:   []string{ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(nameIFace),
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty": {
		loader: loaderTest,
		args:   []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty interface": {
		loader: loaderTest,
		args:   []string{"=Test"},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(&Type{Name: "Test"}),
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty mock": {
		loader: loaderTest,
		args:   []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed mock pattern": {
		loader: loaderTest,
		args: []string{
			"=" + MockPatternDefault,
		},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed match pattern": {
		loader: loaderTest,
		args: []string{
			MatchPatternDefault + "=",
		},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"iface twice different": {
		loader: loaderTest,
		args:   []string{ifaceArg, iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(nameIFace),
			Methods: methodsLoadIFace,
		}, {
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface multiple times": {
		loader: loaderTest,
		args:   []string{ifaceArg, ifaceArg, ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(nameIFace),
			Methods: methodsLoadIFace,
		}},
	},

	"iface failure parsing regexp": {
		loader: loaderTest,
		args: []string{
			"--iface=**",
		},
		expectError: []error{
			NewErrArgFailure(0, "**", NewErrMatcherInvalid(
				targetTest.With(&Type{Name: "**"}),
				&syntax.Error{
					Code: "missing argument to repetition operator",
					Expr: "*",
				})),
		},
	},
	"iface missing": {
		loader: loaderTest,
		args:   []string{"Missing", "Struct", iface},
		expectError: []error{
			NewErrArgFailure(0, "Missing",
				NewErrNotFound(targetTest, "Missing")),
			NewErrArgFailure(1, "Struct",
				NewErrNoIFace(targetTest, "Struct")),
		},
	},

	"failure loading": {
		loader: loaderFail,
		args: []string{
			"--source-file=" + filepath.Join(dirUp, dirMock, dirTest, fileIFace),
		},
		expectError: []error{
			NewErrArgFailure(0, "", NewErrLoading("", fmt.Errorf(
				"err: chdir %s: no such file or directory: stderr: ",
				dirUnknown))),
		},
	},
}

func TestParseMain(t *testing.T) {
	test.Map(t, testParseParams).Run(testParse)
}

var testParseAddParams = map[string]ParseParams{
	"package test": {
		loader: loaderMock,
		args: []string{
			pathTesting, "Test",
			"--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockTestIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test path": {
		loader: loaderMock,
		args: []string{
			dirTesting, "Test",
			"--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockTestIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test file": {
		loader: loaderMock,
		args: []string{
			dirTesting + "/" + fileTesting, "Test",
			"--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockTestIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package gomock": {
		loader: loaderMock,
		args:   []string{pathGoMock, "TestReporter"},
		expectMocks: []*Mock{{
			Source:  sourceGoMockTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "MockTestReporter"}),
			Methods: methodsGoMockTestReporter,
		}},
	},

	"package test and gomock": {
		loader: loaderMock,
		args: []string{
			pathTesting, "Test", "Reporter", pathGoMock, "TestReporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "MockReporter"}),
			Methods: methodsTestReporter,
		}, {
			Source:  sourceGoMockTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "MockTestReporter"}),
			Methods: methodsGoMockTestReporter,
		}},
	},
}

func TestParseAdd(t *testing.T) {
	test.Map(t, testParseAddParams).Run(testParse)
}
