package mock_test

import (
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
	pathGoMock   = "go.uber.org/mock/gomock"

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
		Params: []*Params{{
			Name: "value", Type: "*" + aliasType(mocktest, "Struct"),
		}, {
			Name: "args", Type: "[]*reflect.Value",
		}},
		Results:  []*Params{{Type: "[]any"}, {Type: "error"}},
		Variadic: true,
	}, {
		Name:   "CallB",
		Params: []*Params{},
		Results: []*Params{{
			Name: "fn", Type: "func([]*" + aliasType(mock, "File") + ") []any",
		}, {
			Name: "err", Type: "error",
		}},
	}, {
		Name: "CallC",
		Params: []*Params{{
			Name: "test", Type: aliasType(test, "Context"),
		}},
		Results: []*Params{},
	}}
}

var (
	// Use two different singleton loaders.
	loaderRoot = NewLoader(DirDefault)
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
		Params: []*Params{
			{Name: "cleanup", Type: "func()"},
		},
		Results: []*Params{},
	}, {
		Name:   "Deadline",
		Params: []*Params{},
		Results: []*Params{
			{Name: "deadline", Type: "time.Time"},
			{Name: "ok", Type: "bool"},
		},
	}, {
		Name: "Error",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Errorf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Fail",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name:    "FailNow",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name:    "Failed",
		Params:  []*Params{},
		Results: []*Params{{Type: "bool"}},
	}, {
		Name: "Fatal",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Helper",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name: "Log",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Logf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Name",
		Params:  []*Params{},
		Results: []*Params{{Type: "string"}},
	}, {
		Name:    "Parallel",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name: "Setenv",
		Params: []*Params{
			{Name: "key", Type: "string"},
			{Name: "value", Type: "string"},
		},
		Results: []*Params{},
	}, {
		Name: "Skip",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "SkipNow",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name: "Skipf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Skipped",
		Params:  []*Params{},
		Results: []*Params{{Type: "bool"}},
	}, {
		Name:    "TempDir",
		Params:  []*Params{},
		Results: []*Params{{Type: "string"}},
	}}

	methodsTestReporter = []*Method{{
		Name: "Error",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Errorf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Fail",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name:    "FailNow",
		Params:  []*Params{},
		Results: []*Params{},
	}, {
		Name: "Fatal",
		Params: []*Params{
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name:    "Panic",
		Params:  []*Params{{Name: "arg", Type: "any"}},
		Results: []*Params{},
	}}

	methodsGoMockTestReporter = []*Method{{
		Name: "Errorf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}, {
		Name: "Fatalf",
		Params: []*Params{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Params{},
		Variadic: true,
	}}
)
