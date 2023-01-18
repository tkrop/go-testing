package mock_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

type ParseParams struct {
	root        string
	args        []string
	expectMocks []*Mock
	expectError []error
}

var testParseParams = map[string]ParseParams{
	"no argument": {
		args: []string{},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"invalid argument": {
		args:        []string{"--unknown=any"},
		expectError: []error{NewErrArgInvalid(0, "--unknown=any")},
	},

	"source package explicit": {
		args: []string{"--source=" + pathMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source package derived": {
		args: []string{dirAny},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source package invalid": {
		args: []string{pathUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				pathUnknown, newPackage(pathUnknown))),
		},
	},

	"source directory explicit": {
		args: []string{"--source=" + dirMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source directory derived": {
		args: []string{dirMock},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source directory invalid": {
		root: ".",
		args: []string{dirUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				dirUnknown, newPackage(dirUnknown))),
		},
	},

	"source file explicit": {
		args: []string{"--source=" + fileParser},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source file derived": {
		args: []string{fileParser},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"source file invalid": {
		args: []string{fileUnknown},
		expectError: []error{
			NewErrArgFailure(0, "*", NewErrPackageParsing(
				fileUnknown, newPackage(fileUnknown))),
		},
	},

	"target package explicit": {
		args: []string{"--target-pkg=" + pkgOther},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgOther, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"target package derived": {
		args: []string{"--target=" + pkgOther},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgOther, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"target package guessed": {
		args: []string{pkgOther},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgOther, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},

	"target file explicit": {
		args: []string{"--target-file=" + fileMock},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: Type{
				Package: pkgMock,
				Name:    ifaceMock,
				File:    fileMock,
			},
			Methods: methodsMockIFace,
		}},
	},
	"target file derived": {
		args: []string{"--target=" + fileMock},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: Type{
				Package: pkgMock,
				Name:    ifaceMock,
				File:    fileMock,
			},
			Methods: methodsMockIFace,
		}},
	},
	"target file guessed": {
		args: []string{fileMock},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: Type{
				Package: pkgMock,
				Name:    ifaceMock,
				File:    fileMock,
			},
			Methods: methodsMockIFace,
		}},
	},
	"target file and path guessed": {
		args: []string{dirMock + "/" + fileTarget},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: Type{
				Package: pkgMock,
				Name:    ifaceMock,
				File:    dirMock + "/" + fileTarget,
			},
			Methods: methodsMockIFace,
		}},
	},

	"target file with package guessed": {
		args: []string{fileMock, pkgOther},
		expectMocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: Type{
				Package: pkgOther,
				Name:    ifaceMock,
				File:    fileMock,
			},
			Methods: methodsMockIFace,
		}},
	},

	"iface explicit": {
		args: []string{"--iface=" + iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"iface derived": {
		args: []string{iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"iface with mock": {
		args: []string{ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: iface},
			Methods: methodsMockIFace,
		}},
	},
	"iface with empty mock": {
		args: []string{iface + "="},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"iface with empty interface": {
		args: []string{"=Test"},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: "Test"},
			Methods: methodsMockIFace,
		}},
	},
	"iface twice different": {
		args: []string{ifaceArg, iface},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: iface},
			Methods: methodsMockIFace,
		}, {
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: ifaceMock},
			Methods: methodsMockIFace,
		}},
	},
	"iface multiple times": {
		args: []string{ifaceArg, ifaceArg, ifaceArg},
		expectMocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  Type{Package: pkgMock, Name: iface},
			Methods: methodsMockIFace,
		}},
	},

	"iface failure": {
		args: []string{pathMock, "Missing", "Struct", iface},
		expectError: []error{
			NewErrArgFailure(1, "Missing",
				NewErrNotFound(pathMock, "Missing")),
			NewErrArgFailure(2, "Struct",
				NewErrNoIFace(pathMock, "Struct")),
		},
	},

	"loading failure": {
		root: "nodir", // ensures fallback target
		expectError: []error{
			NewErrArgFailure(-1, "*", NewErrLoading(".", errors.New(
				"err: chdir nodir: no such file or directory: stderr: "))),
		},
	},
}

var testParseAddParams = map[string]ParseParams{
	"package test": {
		args: []string{
			pathTest, "Test", "--target=" + pkgOther, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  Type{Package: pkgMock, Name: "MockTest"},
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  Type{Package: pkgOther, Name: "Reporter"},
			Methods: methodsTestReporter,
		}},
	},

	"package test path": {
		root: ".",
		args: []string{
			dirTest, "Test",
			"--target=" + pkgOther, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  Type{Package: pkgMock, Name: "MockTest"},
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  Type{Package: pkgOther, Name: "Reporter"},
			Methods: methodsTestReporter,
		}},
	},

	"package test file": {
		root: ".",
		args: []string{
			dirTest + "/" + fileTesting, "Test",
			"--target=" + pkgOther, "Reporter=Reporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  Type{Package: pkgMock, Name: "MockTest"},
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  Type{Package: pkgOther, Name: "Reporter"},
			Methods: methodsTestReporter,
		}},
	},

	"package gomock": {
		args: []string{pathGoMock, "TestReporter"},
		expectMocks: []*Mock{{
			Source:  sourceGoMockTestReporter,
			Target:  Type{Package: pkgMock, Name: "MockTestReporter"},
			Methods: methodsGoMockTestReporter,
		}},
	},

	"package test and gomock": {
		args: []string{
			pathTest, "Test", "Reporter", pathGoMock, "TestReporter",
		},
		expectMocks: []*Mock{{
			Source:  sourceTestTest,
			Target:  Type{Package: pkgMock, Name: "MockTest"},
			Methods: methodsTestTest,
		}, {
			Source:  sourceTestReporter,
			Target:  Type{Package: pkgMock, Name: "MockReporter"},
			Methods: methodsTestReporter,
		}, {
			Source:  sourceGoMockTestReporter,
			Target:  Type{Package: pkgMock, Name: "MockTestReporter"},
			Methods: methodsGoMockTestReporter,
		}},
	},
}

func TestParse(t *testing.T) {
	test.Map(t, testParseParams).Run(testParse)
}

func TestParseAdd(t *testing.T) {
	test.Map(t, testParseAddParams).Run(testParse)
}

func testParse(t test.Test, param ParseParams) {
	// Given
	parser := NewParser(loader)
	if param.root != "" {
		loader := NewLoader().(*CachingLoader)
		loader.Config.Dir = param.root
		parser = NewParser(loader)
	}

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
