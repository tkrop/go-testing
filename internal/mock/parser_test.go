package mock_test

import (
	"fmt"
	"path/filepath"
	"regexp/syntax"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
	"golang.org/x/tools/go/packages"
)

var (
	typeFileMock = &Type{File: fileMock}
	typeFileTemp = &Type{File: filepath.Join(dirUp, fileTemplate)}
)

type ParseParams struct {
	loader      Loader
	target      *Type
	args        []string
	expectMocks []*Mock
	expectError []error
}

var parseTestCases = map[string]ParseParams{
	"no argument": {
		loader: loaderTest,
		args:   []string{},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"invalid argument flag": {
		loader:      loaderTest,
		args:        []string{"--test"},
		expectError: []error{NewErrArgInvalid(0, "--test")},
	},
	"invalid argument unknown": {
		loader:      loaderTest,
		args:        []string{"--unknown=any"},
		expectError: []error{NewErrArgInvalid(0, "--unknown=any")},
	},
	// TODO: add test case for invalid argument for guessed type not found.

	"default file": {
		loader: loaderTest,
		args:   []string{},
		target: typeFileMock,
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"default path": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Path: pathTesting},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestingIFace.With(typePkgTestTest),
			Methods: methodsLoadIFace,
		}},
	},
	"default package": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Package: pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"default interface (ignore)": {
		loader: loaderTest,
		args:   []string{},
		target: &Type{Name: ifaceMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"source package explicit": {
		loader: loaderTest,
		args:   []string{"--source-pkg=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source package derived": {
		loader: loaderTest,
		args:   []string{"--source=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source package invalid": {
		loader: NewLoader(DirDefault), // ensure path is not preload.
		args:   []string{pathUnknown},
		expectError: []error{
			NewErrArgFailure(0, ".",
				NewErrPackageParsing(pathUnknown, []*packages.Package{{
					Errors: []packages.Error{{
						Msg: "no required module provides package " + pathUnknown +
							"; to add it:\n\tgo get " + pathUnknown,
					}},
				}}),
			),
		},
	},

	"source path explicit": {
		loader: loaderTest,
		args:   []string{"--source-path=" + pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path derived": {
		loader: loaderTest,
		args:   []string{"--source=" + pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path guessed": {
		loader: loaderRoot,
		args:   []string{pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path default": {
		loader: loaderTest,
		args:   []string{DirDefault},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source path invalid": {
		// Ensure absolute path is pre-loaded for error message.
		loader: loaderMock.PreLoad(absUnknown),
		args:   []string{pathUnknown},
		expectError: []error{
			NewErrArgFailure(0, ".",
				NewErrPackageParsing(absUnknown, []*packages.Package{{
					Errors: []packages.Error{{
						Msg: "stat " + absUnknown + ": directory not found",
					}},
				}}),
			),
		},
	},

	"source file explicit": {
		loader: loaderRoot,
		args:   []string{"--source-file=" + dirSubTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file derived": {
		loader: loaderRoot,
		args:   []string{"--source=" + dirSubTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file derived missing": {
		loader: loaderRoot,
		args: []string{
			"--source=" + fileUnknown,
		},
		expectError: []error{
			NewErrArgNotFound(0, fileUnknown),
		},
		expectMocks: nil,
	},
	"source file guessed": {
		loader: loaderRoot,
		args:   []string{dirSubTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source file guessed invalid": {
		loader: loaderRoot,
		args:   []string{fileUnknown},
		expectError: []error{
			NewErrArgFailure(0, ".",
				NewErrPackageParsing(fileUnknown, []*packages.Package{{
					Errors: []packages.Error{{
						Msg: "no required module provides package " + fileUnknown +
							"; to add it:\n\tgo get " + fileUnknown,
					}},
				}}),
			),
		},
	},

	"source directory explicit": {
		loader: loaderRoot,
		args:   []string{"--source-file=" + dirSubTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory derived": {
		loader: loaderRoot,
		args:   []string{"--source=" + dirSubTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory guessed": {
		loader: loaderRoot,
		args:   []string{dirSubTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"source directory invalid": {
		loader: loaderRoot,
		args:   []string{dirUnknown},
		expectError: []error{
			NewErrArgFailure(0, ".",
				NewErrPackageParsing(dirUnknown, []*packages.Package{{
					Errors: []packages.Error{{
						Msg: "stat " + absUnknown + ": directory not found",
					}},
				}}),
			),
		},
	},

	"target package explicit": {
		loader: loaderTest,
		args:   []string{"--target-pkg=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"target package derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"target package guessed": {
		loader: loaderTest,
		args:   []string{pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"target path explicit": {
		loader: loaderTest,
		args:   []string{"--target-path=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"target path derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"target file explicit": {
		loader: loaderTest,
		args:   []string{"--target-file=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file derived": {
		loader: loaderTest,
		args:   []string{"--target=" + typeFileTemp.File},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace.With(typeFileTemp),
			Methods: methodsLoadIFace,
		}},
	},
	"target file derived missing": {
		loader: loaderTest,
		args:   []string{"--target=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file guessed": {
		loader: loaderTest,
		args:   []string{fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},
	"target file with path guessed": {
		loader: loaderTest,
		args:   []string{filepath.Join(dirUp, fileMock)},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetMockTestIFace.With(&Type{
				File: filepath.Join(dirUp, fileMock),
			}),
			Methods: methodsLoadIFace,
		}},
	},

	"target file with package guessed": {
		loader: loaderTest,
		args:   []string{fileMock, pkgTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestIFace.With(typeFileMock),
			Methods: methodsLoadIFace,
		}},
	},

	"iface explicit": {
		loader: loaderTest,
		args:   []string{"--iface=" + iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed": {
		loader: loaderTest,
		args:   []string{iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed with mock": {
		loader: loaderTest,
		args:   []string{ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(nameIFace),
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty": {
		loader: loaderTest,
		args:   []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty interface": {
		loader: loaderTest,
		args:   []string{"=Test"},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(&Type{Name: "Test"}),
			Methods: methodsLoadIFace,
		}},
	},
	"iface guessed empty mock": {
		loader: loaderTest,
		args:   []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
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
			Target:  targetTestTestIFace,
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
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},

	"iface twice different": {
		loader: loaderTest,
		args:   []string{ifaceArg, iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(nameIFace),
			Methods: methodsLoadIFace,
		}, {
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace,
			Methods: methodsLoadIFace,
		}},
	},
	"iface multiple times": {
		loader: loaderTest,
		args:   []string{ifaceArg, ifaceArg, ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetTestTestIFace.With(nameIFace),
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
			"--source-file=" + filepath.Join(dirUp, dirMock, dirSubTest, fileIFace),
		},
		expectError: []error{
			NewErrArgFailure(0, ".", NewErrLoading("", fmt.Errorf(
				"err: chdir %s: no such file or directory: stderr: ",
				dirUnknown))),
		},
	},
}

var parseAddTestCases = map[string]ParseParams{
	"package test": {
		loader: loaderRoot,
		args: []string{
			pathTesting, "Test",
			"--target=" + pkgMock, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockTestIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test path": {
		loader: loaderRoot,
		args: []string{
			dirTest, "Test",
			"--target=" + pkgMock, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockTestIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test file": {
		loader: loaderRoot,
		args: []string{
			dirTest + "/" + fileContext, "Test",
			"--target=" + pkgMock, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockTestIFace.With(&Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFace.With(&Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package gomock": {
		loader: loaderRoot,
		args:   []string{pathGoMock, "TestReporter"},
		expectMocks: []*Mock{{
			Source:  sourceGoMockTestReporter,
			Target:  targetMockTestIFace.With(&Type{Name: "MockTestReporter"}),
			Methods: methodsGoMockTestReporter,
		}},
	},

	"package test and gomock": {
		loader: loaderRoot,
		args: []string{
			pkgMock,
			pathTesting, "Test", "Reporter",
			pathGoMock, "TestReporter",
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

func testParse(t test.Test, param ParseParams) {
	// Given
	parser := NewParser(param.loader, param.target)

	// When
	mocks, errs := parser.Parse(param.args...)

	// Then
	assert.Equal(t, param.expectError, errs)
	assert.Equal(t, param.expectMocks, mocks)
}

func TestParseMain(t *testing.T) {
	test.Map(t, parseTestCases).Run(testParse)
}

func TestParseAdd(t *testing.T) {
	test.Map(t, parseAddTestCases).Run(testParse)
}
