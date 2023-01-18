package mock_test

import (
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	. "github.com/tkrop/go-testing/internal/mock"
)

const (
	pkgMock   = "mock"
	pkgTest   = "test"
	pkgGoMock = "gomock"
	pkgOther  = "other"

	dirAny     = "."
	dirMock    = "../mock"
	dirUnknown = "../unknown"
	dirTest    = "../../test"

	pathMock    = "github.com/tkrop/go-testing/internal/mock/mock"
	pathUnknown = "github.com/tkrop/go-testing/internal/unknown"
	pathTest    = "github.com/tkrop/go-testing/test"
	pathGoMock  = "github.com/golang/mock/gomock"

	fileMock    = "mock_test.go"
	fileUnknown = "unnkown_test.go"
	fileParser  = "parser_test.go"
	fileTarget  = "mock_parser.go"
	fileTesting = "testing.go"

	ifaceAny     = "*"
	ifaceAnyMock = "Mock*"
	iface        = "IFace"
	ifaceMock    = "MockIFace"
	ifaceArg     = iface + "=" + iface
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

var (
	// Use singleton caching loader with switched dir to speedup tests.
	loader = func() Loader {
		loader := NewLoader().(*CachingLoader)
		loader.Config.Dir = "./mock"
		return loader
	}()

	pkgParsedMock   = newPackage(pathMock)[1]
	pkgParsedTest   = newPackage(pathTest)[0]
	pkgParsedGoMock = newPackage(pathGoMock)[0]

	sourceIFaceAny = Type{
		Path: pathMock, File: getMethod(pkgParsedMock, iface),
		Package: pkgMock, Name: iface,
	}
	sourceGoMockTestReporter = Type{
		Path: pathGoMock, File: getMethod(pkgParsedGoMock, "TestReporter"),
		Package: pkgGoMock, Name: "TestReporter",
	}
	sourceTestTest = Type{
		Path: pathTest, File: getMethod(pkgParsedTest, "Test"),
		Package: pkgTest, Name: "Test",
	}
	sourceTestReporter = Type{
		Path: pathTest, File: getMethod(pkgParsedTest, "Reporter"),
		Package: pkgTest, Name: "Reporter",
	}

	methodsMockIFace = []*Method{{
		Name: "Call",
		Params: []*Param{
			{Name: "value", Type: "*" + pathMock + ".Struct"},
			{Name: "args", Type: "[]*reflect.Value"},
		},
		Results:  []*Param{{Type: "[]any"}},
		Variadic: true,
	}}

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
