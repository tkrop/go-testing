package mock_test

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	. "github.com/tkrop/go-testing/internal/mock"
)

const (
	pkgMock         = "mock"
	pkgMockTest     = "mock_test"
	pkgTest         = "test"
	pkgTestTest     = "test_test"
	pkgUnknown      = "unknown"
	pkgUnknownTest  = "unknown_test"
	pkgInternal     = "internal"
	pkgInternalTest = "internal_test"
	pkgTesting      = "testing_test"
	pkgGoMock       = "gomock"

	dirUp   = ".."
	dirTop  = "../../.."
	dirMock = "mock"

	dirSubTest = "./test"
	dirOther   = "../other"
	dirUnknown = "../unknown"
	dirTest    = "../../test"

	pathMock     = "github.com/tkrop/go-testing/internal/mock"
	pathTest     = "github.com/tkrop/go-testing/internal/mock/test"
	pathUnknown  = "github.com/tkrop/go-testing/internal/unknown"
	pathInternal = "github.com/tkrop/go-testing/internal"
	pathTesting  = "github.com/tkrop/go-testing/test"
	pathGoMock   = "github.com/golang/mock/gomock"

	fileIFace    = "iface.go"
	fileMock     = "mock_test.go"
	fileOther    = "mock_other_test.go"
	fileTemplate = "mock_template_test.go"
	fileUnknown  = "unnkown_test.go"
	fileContext  = "context.go"

	aliasMock   = "mock_" + pkgTest
	aliasInt    = "internal_" + aliasMock
	aliasRepo   = "testing_" + aliasInt
	aliasOrg    = "tkrop_" + aliasRepo
	aliasCom    = "com_" + aliasOrg
	aliasGitHub = "github_" + aliasCom

	iface     = "IFace"
	ifaceMock = "MockIFace"
	ifaceArg  = iface + "=" + iface
)

var (
	errAny = errors.New("any error")

	absUnknown, _ = filepath.Abs(dirUnknown)

	nameIFace     = &Type{Name: iface}
	nameIFaceMock = &Type{Name: ifaceMock}

	typePkgTest     = &Type{Package: pkgTest}
	typePkgTestTest = &Type{Package: pkgTestTest}

	targetTest         = &Type{Package: pkgTest, Path: pathTest}
	targetTestTest     = &Type{Package: pkgTestTest, Path: pathTest}
	targetTesting      = &Type{Package: pkgTest, Path: pathTesting}
	targetMock         = &Type{Package: pkgMock, Path: pathMock}
	targetMockTest     = &Type{Package: pkgMockTest, Path: pathMock}
	targetUnknown      = &Type{Package: pkgUnknown, Path: pathUnknown}
	targetUnknownTest  = &Type{Package: pkgUnknownTest, Path: pathUnknown}
	targetInternal     = &Type{Package: pkgInternal, Path: pathInternal}
	targetInternalTest = &Type{Package: pkgInternalTest, Path: pathInternal}

	targetTestIFace     = targetTest.With(nameIFaceMock)
	targetTestTestIFace = targetTestTest.With(nameIFaceMock)

	targetMockIFace     = targetMock.With(nameIFaceMock)
	targetMockTestIFace = targetMockTest.With(nameIFaceMock)
	targetTestingIFace  = targetTesting.With(nameIFaceMock)
)

func newPackage(path string) []*packages.Package {
	pkgs, _ := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedTypes,
		Tests: true,
	}, path)
	return pkgs
}

func getMethod(pkg *packages.Package, name string) string {
	pos := pkg.Fset.Position(pkg.Types.Scope().Lookup(name).Pos())
	return fmt.Sprintf("%s:%d", filepath.Base(pos.Filename), pos.Line)
}

func aliasType(alias, stype string) string {
	if alias != "" {
		return alias + "." + stype
	}
	return stype
}

func methodsMockIFaceFunc(mocktest, test, mock string) []*Method {
	return []*Method{{
		Name: "CallA",
		Params: []*Param{{
			Name: "value", Type: "*" + aliasType(mocktest, "Struct"),
		}, {
			Name: "args", Type: "[]*reflect.Value",
		}},
		Results:  []*Param{{Type: "[]any"}, {Type: "error"}},
		Variadic: true,
	}, {
		Name:   "CallB",
		Params: []*Param{},
		Results: []*Param{{
			Name: "fn", Type: "func([]*" + aliasType(mock, "File") + ") []any",
		}, {
			Name: "err", Type: "error",
		}},
	}, {
		Name: "CallC",
		Params: []*Param{{
			Name: "test", Type: aliasType(test, "Context"),
		}},
		Results: []*Param{},
	}}
}

var (
	// Use two different singleton loaders.
	loaderMock = NewLoader(DirDefault)
	loaderTest = NewLoader(dirSubTest)
	loaderFail = NewLoader(dirUnknown)

	// Use singleton template for testing.
	template = MustTemplate()

	pkgParsedMock   = newPackage(pathTest)[0]
	pkgParsedTest   = newPackage(pathTesting)[0]
	pkgParsedGoMock = newPackage(pathGoMock)[0]

	sourceIFaceAny = &Type{
		Path: pathTest, File: getMethod(pkgParsedMock, iface),
		Package: pkgTest, Name: iface,
	}
	sourceGoMockTestReporter = &Type{
		Path: pathGoMock, Package: pkgGoMock,
		File: getMethod(pkgParsedGoMock, "TestReporter"),
		Name: "TestReporter",
	}
	sourceTestTest = &Type{
		Package: pkgTest, Path: pathTesting,
		File: getMethod(pkgParsedTest, "Test"),
		Name: "Test",
	}
	sourceTestReporter = &Type{
		Path: pathTesting, Package: pkgTest,
		File: getMethod(pkgParsedTest, "Reporter"),
		Name: "Reporter",
	}

	methodsLoadIFace = methodsMockIFaceFunc(
		pathTest, pathTesting, pathMock)

	methodsTestTest = []*Method{{
		Name: "Cleanup",
		Params: []*Param{
			{Name: "cleanup", Type: "func()"},
		},
		Results: []*Param{},
	}, {
		Name:   "Deadline",
		Params: []*Param{},
		Results: []*Param{
			{Name: "deadline", Type: "time.Time"},
			{Name: "ok", Type: "bool"},
		},
	}, {
		Name: "Error",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Errorf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Fail",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name:    "FailNow",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name:    "Failed",
		Params:  []*Param{},
		Results: []*Param{{Type: "bool"}},
	}, {
		Name: "Fatal",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Helper",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name: "Log",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Logf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Name",
		Params:  []*Param{},
		Results: []*Param{{Type: "string"}},
	}, {
		Name:    "Parallel",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name: "Setenv",
		Params: []*Param{
			{Name: "key", Type: "string"},
			{Name: "value", Type: "string"},
		},
		Results: []*Param{},
	}, {
		Name: "Skip",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "SkipNow",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name: "Skipf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Skipped",
		Params:  []*Param{},
		Results: []*Param{{Type: "bool"}},
	}, {
		Name:    "TempDir",
		Params:  []*Param{},
		Results: []*Param{{Type: "string"}},
	}}

	methodsTestReporter = []*Method{{
		Name: "Error",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Errorf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Fail",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name:    "FailNow",
		Params:  []*Param{},
		Results: []*Param{},
	}, {
		Name: "Fatal",
		Params: []*Param{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:    "Panic",
		Params:  []*Param{{Name: "arg", Type: "any"}},
		Results: []*Param{},
	}}

	methodsGoMockTestReporter = []*Method{{
		Name: "Errorf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}}
)
