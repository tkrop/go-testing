package mock_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

type ParseParams struct {
	loader      Loader
	target      Type
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
	if param.expectError != nil {
		assert.Equal(t, param.expectError, errs)
	} else {
		assert.Empty(t, errs)
	}
	assert.Equal(t, param.expectMocks, mocks)
}

var testParseParams = map[string]ParseParams{
	"no argument": {
		loader: loaderTest,
		args:   []string{},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
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
		target: Type{File: fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{File: fileMock}),
			Methods: methodsMockIFace,
		}},
	},
	"default path (ignored)": {
		loader: loaderTest,
		args:   []string{},
		target: Type{Path: pathTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"default package": {
		loader: loaderTest,
		args:   []string{},
		target: Type{Package: pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest,
			Methods: methodsMockIFace,
		}},
	},
	"default interface (ignore)": {
		loader: loaderTest,
		args:   []string{},
		target: Type{Name: iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},

	"source package explicit": {
		loader: loaderTest,
		args:   []string{"--source=" + pathMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source package derived": {
		loader: loaderTest,
		args:   []string{pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest,
			Methods: methodsMockIFace,
		}},
	},
	"source package default": {
		loader: loaderTest,
		args:   []string{DefaultDir},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source package invalid": {
		loader: loaderMock,
		args:   []string{pathUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				pathUnknown, newPackage(pathUnknown))),
		},
	},

	"source directory explicit": {
		loader: loaderMock,
		args:   []string{"--source=" + dirMockTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source directory derived": {
		loader: loaderMock,
		args:   []string{dirMockTest + "/" + fileIFace},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source directory invalid": {
		loader: loaderMock,
		args:   []string{dirUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				dirUnknown, newPackage(dirUnknown))),
		},
	},

	"source file explicit": {
		loader: loaderMock,
		args:   []string{"--source=" + dirMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source file derived": {
		loader: loaderMock,
		args:   []string{dirMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"source file invalid": {
		loader: loaderMock,
		args:   []string{dirUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				dirUnknown, newPackage(dirUnknown))),
		},
	},

	"target package explicit": {
		loader: loaderTest,
		args:   []string{"--target-pkg=" + pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest,
			Methods: methodsMockIFace,
		}},
	},
	"target package derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest,
			Methods: methodsMockIFace,
		}},
	},
	"target package guessed": {
		loader: loaderTest,
		args:   []string{pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest,
			Methods: methodsMockIFace,
		}},
	},

	"target path explicit": {
		loader: loaderTest,
		args:   []string{"--target-path=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"target path derived": {
		loader: loaderTest,
		args:   []string{"--target=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},

	"target file explicit": {
		loader: loaderTest,
		args:   []string{"--target-file=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{File: fileMock}),
			Methods: methodsMockIFace,
		}},
	},
	"target file derived": {
		loader: loaderTest,
		args:   []string{"--target=" + fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{File: fileMock}),
			Methods: methodsMockIFace,
		}},
	},
	"target file guessed": {
		loader: loaderTest,
		args:   []string{fileMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{File: fileMock}),
			Methods: methodsMockIFace,
		}},
	},
	"target file with path guessed": {
		loader: loaderTest,
		args:   []string{dirMock + "/" + fileMock},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetMockIFace.With(Type{
				File: dirMock + "/" + fileMock,
			}),
			Methods: methodsMockIFace,
		}},
	},

	"target file with package guessed": {
		loader: loaderTest,
		args:   []string{fileMock, pkgMockTest},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFaceTest.With(Type{File: fileMock}),
			Methods: methodsMockIFace,
		}},
	},

	"iface explicit": {
		loader: loaderTest,
		args:   []string{"--iface=" + iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"iface derived": {
		loader: loaderTest,
		args:   []string{iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"iface with mock": {
		loader: loaderTest,
		args:   []string{ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{Name: iface}),
			Methods: methodsMockIFace,
		}},
	},
	"iface with empty mock": {
		loader: loaderTest,
		args:   []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"iface with empty interface": {
		loader: loaderTest,
		args:   []string{"=Test"},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{Name: "Test"}),
			Methods: methodsMockIFace,
		}},
	},
	"iface twice different": {
		loader: loaderTest,
		args:   []string{ifaceArg, iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{Name: iface}),
			Methods: methodsMockIFace,
		}, {
			Source:  sourceIFaceAny,
			Target:  targetMockIFace,
			Methods: methodsMockIFace,
		}},
	},
	"iface multiple times": {
		loader: loaderTest,
		args:   []string{ifaceArg, ifaceArg, ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{Name: iface}),
			Methods: methodsMockIFace,
		}},
	},

	"iface failure": {
		loader: loaderTest,
		args:   []string{pathMockTest, "Missing", "Struct", iface},
		expectError: []error{
			NewErrArgFailure(1, "Missing",
				NewErrNotFound(pathMockTest, "Missing")),
			NewErrArgFailure(2, "Struct",
				NewErrNoIFace(pathMockTest, "Struct")),
		},
	},

	"loading failure": {
		// ensures package loading failure.
		loader: NewLoader(dirUnknown),
		expectError: []error{
			NewErrArgFailure(-1, "*", NewErrLoading(DefaultDir, fmt.Errorf(
				"err: chdir %s: no such file or directory: stderr: ",
				dirUnknown))),
		},
	},
}

// TODO: var testParseXParams = map[string]ParseParams{}

func TestParseMain(t *testing.T) {
	test.Map(t, testParseParams).Run(testParse)
}

var testParseAddParams = map[string]ParseParams{
	"package test": {
		loader: loaderMock,
		args: []string{
			pathTest, "Test", "--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFaceTest.With(Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test path": {
		loader: loaderMock,
		args: []string{
			dirTest, "Test",
			"--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFaceTest.With(Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package test file": {
		loader: loaderMock,
		args: []string{
			dirTest + "/" + fileTesting, "Test",
			"--target=" + pkgMockTest, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFaceTest.With(Type{Name: "Reporter"}),
			Methods: methodsTestReporter,
		}},
	},

	"package gomock": {
		loader: loaderMock,
		args:   []string{pathGoMock, "TestReporter"},
		expectMocks: []*Mock{{
			Source:  sourceGoMockTestReporter,
			Target:  targetMockIFace.With(Type{Name: "MockTestReporter"}),
			Methods: methodsGoMockTestReporter,
		}},
	},

	"package test and gomock": {
		loader: loaderMock,
		args: []string{
			pathTest, "Test", "Reporter", pathGoMock, "TestReporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  targetMockIFace.With(Type{Name: "MockTest"}),
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  targetMockIFace.With(Type{Name: "MockReporter"}),
			Methods: methodsTestReporter,
		}, {
			Source:  sourceGoMockTestReporter,
			Target:  targetMockIFace.With(Type{Name: "MockTestReporter"}),
			Methods: methodsGoMockTestReporter,
		}},
	},
}

func TestParseAdd(t *testing.T) {
	test.Map(t, testParseAddParams).Run(testParse)
}
