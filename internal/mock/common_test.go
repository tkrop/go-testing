package mock_test

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	. "github.com/tkrop/go-testing/internal/mock"
)

const (
	pkgMock     = "mock"
	pkgMockTest = "mock_test"
	pkgTest     = "test"
	pkgTesting  = "testing_test"
	pkgGoMock   = "gomock"

	dirMock     = "../mock"
	dirMockTest = "../mock/test"
	dirOther    = "../other"
	dirUnknown  = "../unknown"
	dirTest     = "../../test"

	pathMock     = "github.com/tkrop/go-testing/internal/mock"
	pathMockTest = "github.com/tkrop/go-testing/internal/mock/test"
	pathUnknown  = "github.com/tkrop/go-testing/internal/unknown"
	pathTest     = "github.com/tkrop/go-testing/test"
	pathGoMock   = "github.com/golang/mock/gomock"

	fileIFace   = "iface.go"
	fileMock    = "mock_test.go"
	fileOther   = "mock_other_test.go"
	fileUnknown = "unnkown_test.go"
	fileTesting = "testing.go"

	aliasMock   = "mock_" + pkgTest
	aliasInt    = "internal_" + aliasMock
	aliasRepo   = "testing_" + aliasInt
	aliasOrg    = "tkrop_" + aliasRepo
	aliasCom    = "com_" + aliasOrg
	aliasGitHub = "github_" + aliasCom

	ifaceAny     = "*"
	ifaceAnyMock = "Mock*"
	iface        = "IFace"
	ifaceMock    = "MockIFace"
	ifaceArg     = iface + "=" + iface
)

var (
	errAny = errors.New("any error")

	targetMockIFace = Type{
		Package: pkgMock, Path: pathMock, Name: ifaceMock,
	}
	targetMockIFaceTest = Type{
		Package: pkgMockTest, Path: pathMock, Name: ifaceMock,
	}
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
		Variadic: false,
	}, {
		Name: "CallC",
		Params: []*Param{{
			Name: "test", Type: aliasType(test, "Tester"),
		}},
		Results:  []*Param{},
		Variadic: false,
	}}
}

var (
	// Use two different singleton loaders.
	loaderMock = NewLoader(".")
	loaderTest = NewLoader(dirMockTest)

	// Use singleton template for testing.
	template = MustTemplate()

	pkgParsedMock   = newPackage(pathMockTest)[0]
	pkgParsedTest   = newPackage(pathTest)[0]
	pkgParsedGoMock = newPackage(pathGoMock)[0]

	sourceIFaceAny = Type{
		Path: pathMockTest, File: getMethod(pkgParsedMock, iface),
		Package: pkgTest, Name: iface,
	}
	sourceGoMockTestReporter = Type{
		Path: pathGoMock, Package: pkgGoMock,
		File: getMethod(pkgParsedGoMock, "TestReporter"),
		Name: "TestReporter",
	}
	sourceTestTest = Type{
		Package: pkgTest, Path: pathTest,
		File: getMethod(pkgParsedTest, "Test"),
		Name: "Test",
	}
	sourceTestReporter = Type{
		Path: pathTest, Package: pkgTest,
		File: getMethod(pkgParsedTest, "Reporter"),
		Name: "Reporter",
	}

	methodsMockIFace = methodsMockIFaceFunc(
		pathMockTest, pathTest, pathMock)

	methodsTestTest = []*Method{{
		Name: "Errorf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:     "FailNow",
		Params:   []*Param{},
		Results:  []*Param{},
		Variadic: false,
	}, {
		Name: "Fatalf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:     "Helper",
		Params:   []*Param{},
		Results:  []*Param{},
		Variadic: false,
	}, {
		Name:     "Name",
		Params:   []*Param{},
		Results:  []*Param{{Type: "string"}},
		Variadic: false,
	}}

	methodsTestReporter = []*Method{{
		Name: "Errorf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:     "FailNow",
		Params:   []*Param{},
		Results:  []*Param{},
		Variadic: false,
	}, {
		Name: "Fatalf",
		Params: []*Param{
			{Name: "format", Type: "string"},
			{Name: "args", Type: "[]any"},
		},
		Results:  []*Param{},
		Variadic: true,
	}, {
		Name:     "Panic",
		Params:   []*Param{{Name: "arg", Type: "any"}},
		Results:  []*Param{},
		Variadic: false,
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
