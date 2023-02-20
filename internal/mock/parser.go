package mock

import (
	"errors"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

const (
	// Default directory for package loading.
	DefaultDir = "."
)

// Parser is a mock setup command line argument parser.
type Parser struct {
	// A caching, reusable package loader.
	loader Loader
	// The default target definition for parsing.
	target Type
}

// NewParser creates a new mock setup command line argument parser with loading
// of test packages enabled or disabled.
func NewParser(loader Loader, target Type) *Parser {
	return &Parser{loader: loader, target: target}
}

// parse parses the command line arguments provided and returns a list of mock
// stubs containing the necessary source and target information without method
// data.
func (parser *Parser) Parse(args ...string) ([]*Mock, []error) {
	state := newParseState(parser)
	for pos, arg := range args {
		switch atype, arg := state.parseArgType(arg); atype {
		case argTypeSource:
			state.ensureIFace(pos)
			state.source = Type{File: arg}
		case argTypePackage:
			state.target.Package = arg
		case argTypePath:
			state.target.Path = arg
		case argTypeFile:
			state.target.File = arg
		case argTypeIFace:
			for _, arg := range strings.Split(arg, ",") {
				state.creatMocks(pos, arg)
			}
		default:
			err := NewErrArgInvalid(pos, arg)
			state.errs = append(state.errs, err)
		}
	}

	state.ensureSource().
		ensureIFace(len(args) - 1)
	if len(state.errs) != 0 {
		return nil, state.errs
	}
	return state.mocks, nil
}

// argType defines the command line argument types.
type argType int

// Collection of command line argument types.
const (
	// Unknown argument type that results in an error.
	argTypeUnknown argType = iota
	// Target package argument type (must be identifier).
	argTypePackage
	// Target file argument type (must end with `.go`).
	argTypePath
	// Target import path argument type (must resolve to a loadable package).
	argTypeFile
	// Source package, directory, or file argument type (very flexible).
	argTypeSource
	// Source and target interface argument type (must be list of identifier
	// mappings).
	argTypeIFace
)

// ParseState is keeping the state during the parsing process.
type ParseState struct {
	// Loader instance for package loader.
	loader Loader
	// Mape of target to mock interface configurations.
	targets map[Type]*Mock
	// Collected mock interfaces configurations.
	mocks []*Mock
	// Collected errors during parsing and package loading.
	errs []error
	// Source provides the actual source interface setup.
	source Type
	// Target provides the actual target interface setup.
	target Type
}

// newParseState creates a new parse state for parsing.
func newParseState(parser *Parser) *ParseState {
	return &ParseState{
		loader:  parser.loader,
		source:  Type{},
		target:  parser.target,
		targets: map[Type]*Mock{},
		mocks:   []*Mock{},
		errs:    []error{},
	}
}

// parseArgType parser the actual command line argument type from the command
// line argument and returns the argument type together with the remaining
// argument.
func (state *ParseState) parseArgType(arg string) (argType, string) {
	if strings.Index(arg, "--") == 0 {
		equal := strings.Index(arg, "=")
		flag, sarg := arg[2:equal], arg[equal+1:]
		switch flag {
		case "iface":
			return argTypeIFace, sarg
		case "source":
			return argTypeSource, sarg
		case "target":
			if token.IsIdentifier(sarg) {
				return argTypePackage, sarg
			}
			pkgs, err := state.loader.LoadPackage(sarg)
			if len(pkgs) > 0 && err == nil {
				return argTypePath, sarg
			}
			return argTypeFile, sarg
		case "target-pkg":
			return argTypePackage, sarg
		case "target-path":
			return argTypePath, sarg
		case "target-file":
			return argTypeFile, sarg
		default:
			return argTypeUnknown, arg
		}
	}

	if strings.ContainsAny(arg, "=,") {
		return argTypeIFace, arg
	} else if token.IsIdentifier(arg) {
		if token.IsExported(arg) {
			return argTypeIFace, arg
		}
		return argTypePackage, arg
	}

	if strings.HasSuffix(arg, ".go") {
		file := arg
		if index := strings.LastIndex(arg, "/"); index >= 0 {
			file = arg[index+1:]
		}
		if strings.HasPrefix(file, "mock_") {
			return argTypeFile, arg
		}
	}
	return argTypeSource, arg
}

// ensureSource checks the current source setup and derives the default package
// name and package path from the given source file, if the setup is missing.
// The default name is ensured when parsing the interface specific command line
// arguments.
func (state *ParseState) ensureSource() *ParseState {
	if state.source.File == "" {
		state.source.File = DefaultDir
	}
	return state.ensureType(&state.source)
}

// ensureTarget checks the current target setup and derives the default package
// name and package path from the given target file, if the setup is missing.
// The default name is ensured when parsing the interface specific command line
// arguments.
func (state *ParseState) ensureTarget() *ParseState {
	return state.ensureType(&state.target)
}

// ensureType checks the current setup and derives the default package name and
// package path from the given file, if the setup is missing. The default name
// is ensured when parsing the interface command line arguments.
func (state *ParseState) ensureType(ptype *Type) *ParseState {
	// found no way to make filepath.Abs to provide an error.
	file, _ := filepath.Abs(ptype.File)
	if info, err := os.Stat(file); err != nil || !info.IsDir() {
		file = filepath.Dir(file)
	}

	if pkgs, err := state.loader.LoadPackage(file); err == nil {
		ptype.Path = pkgs[0].PkgPath
		// TODO: support only valid package names!
		if ptype.Package == "" {
			ptype.Package = pkgs[0].Name
		}
	}
	return state
}

// ensuresArg checks and interface command line argument and if incomplete
// extends it with the default interface and interfac mock names.
func (state *ParseState) ensuresArg(arg string) string {
	if arg == "" {
		state.source.Name = "*"
		state.target.Name = "Mock*"
		return "*"
	}

	index := strings.IndexAny(arg, "=")
	switch index {
	case -1:
		state.source.Name = arg
		state.target.Name = "Mock" + arg
	case 0:
		state.source.Name = "*"
		state.target.Name = arg[1:]
	case len(arg) - 1:
		state.source.Name = arg[:index]
		state.target.Name = "Mock*"
	default:
		state.source.Name = arg[:index]
		state.target.Name = arg[index+1:]
	}
	return arg
}

// ensureIFace ensures that the last source package was used to create at least
// oone mock stub before we continue with the next source package. If now mock
// stub was created, a generic mock stub is created that mocks all interfaces
// in this package.
func (state *ParseState) ensureIFace(pos int) {
	if state.source.File != "" && state.source.Name == "" {
		state.creatMocks(pos, "")
	}
}

// creatMocks creates the mock stubs for the given source/target interface
// command line argument. If no target interface name/pattern is provided it
// is set to the source interface name prefixed by `Mock`. If no source
// interface name/pattern is provided it is assumed to be any interface in the
// requested package.
func (state *ParseState) creatMocks(pos int, arg string) {
	arg = state.ensureSource().ensureTarget().ensuresArg(arg)

	ifaces, err := state.loader.LoadIFaces(state.source)
	if err != nil {
		state.errs = append(state.errs, NewErrArgFailure(pos, arg, err))
	}

	for _, iface := range ifaces {
		target := state.target
		target.Name = strings.ReplaceAll(target.Name, "*", iface.Source.Name)
		if _, ok := state.targets[target]; !ok {
			mock := &Mock{
				Source:  iface.Source,
				Target:  target,
				Methods: iface.Methods,
			}
			state.targets[target] = mock
			state.mocks = append(state.mocks, mock)
		}
	}
}

// Loader is the generic interface of the package and interface loader.
type Loader interface {
	// LoadPackage is loading the go-packages for the given path. The details
	// depend on the backing package loader configuration.
	LoadPackage(path string) ([]*packages.Package, error)
	// LoadIFaces is loading the go-packages for the given source and extracts
	// the interfaces matching the interface naming pattern.
	LoadIFaces(source Type) ([]*IFace, error)
}

// CachingLoader allows to efficient load, parse, and analyze packages as well
// as interfaces. The loaded packages are chached by request path for repeated
// access and analyzing. The loader is safe to be used concurrently.
type CachingLoader struct {
	// Configuration for package loading.
	Config *packages.Config
	// Cache for packages mapped to load paths.
	pkgs map[string][]*packages.Package
	// Mutext to support concurrent usage.
	mutex sync.Mutex
}

// NewLoader creates a new caching package loader, that allows efficient access
// to types from files and packages. By default it provides also access to the
// test packages and searches these for interfaces. To disable you need to set
// `loader.(*CachingLoader).Config.Tests` to `false`.
func NewLoader(dir string) Loader {
	return &CachingLoader{
		Config: &packages.Config{
			Mode: packages.NeedName |
				packages.NeedTypes |
				packages.NeedSyntax,
			Tests: true,
			Dir:   dir,
		},
		pkgs: map[string][]*packages.Package{},
	}
}

// LoadPackage is loading the go-packages for the given path and caching them
// for repeated requests. The details provided by the package information
// depend on the package loader configuration. By default only types and names
// are loaded.
func (loader *CachingLoader) LoadPackage(
	path string,
) ([]*packages.Package, error) {
	return loader.lookupPackage(path)
}

// LoadIFaces is loading the go-packages for the given source and extracts the
// interfaces matching the interface naming pattern.
func (loader *CachingLoader) LoadIFaces(source Type) ([]*IFace, error) {
	pkgs, err := loader.lookupPackage(source.File)
	if err != nil {
		return nil, err
	} else if source.Name == "*" {
		return loader.lookupIFacesAny(pkgs, source)
	}
	return loader.lockupIFacesNamed(pkgs, source)
}

func (loader *CachingLoader) lookupPackage(
	path string,
) ([]*packages.Package, error) {
	// Ensures loading and storing of packages, not files.
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}

	loader.mutex.Lock()
	if pkgs, ok := loader.pkgs[path]; ok {
		loader.mutex.Unlock()
		return pkgs, nil
	}
	loader.mutex.Unlock()

	if pkgs, err := packages.Load(loader.Config, path); err != nil {
		return nil, NewErrLoading(path, err)
	} else if err := NewErrPackageParsing(path, pkgs); err != nil {
		return nil, err
	} else {
		loader.mutex.Lock()
		loader.pkgs[path] = pkgs
		loader.mutex.Unlock()
		return pkgs, nil
	}
}

func (loader *CachingLoader) lookupIFace(
	pkg *packages.Package, name string,
) (*types.TypeName, *types.Interface, error) {
	if object := pkg.Types.Scope().Lookup(name); object == nil {
		return nil, nil, NewErrNotFound(pkg.PkgPath, name)
	} else if iface, ok := object.Type().Underlying().(*types.Interface); !ok {
		return nil, nil, NewErrNoIFace(pkg.PkgPath, name)
	} else if named, ok := object.(*types.TypeName); !ok {
		return nil, nil, NewErrNoNameType(pkg.PkgPath, name)
	} else {
		return named, iface, nil
	}
}

func (loader *CachingLoader) lockupIFacesNamed(
	pkgs []*packages.Package, source Type,
) ([]*IFace, error) {
	ifaces := []*IFace{}
	for _, pkg := range pkgs {
		name, iface, err := loader.lookupIFace(pkg, source.Name)
		if err == nil {
			ifaces = append(ifaces, NewIFace(name, pkg.Fset, iface))
		} else if errors.Is(err, ErrNoIFace) {
			return nil, err
		}
	}

	if len(ifaces) == 0 {
		return ifaces, NewErrNotFound(source.File, source.Name)
	}
	return ifaces, nil
}

func (loader *CachingLoader) lookupIFacesAny(
	pkgs []*packages.Package, source Type, //nolint:unparam
) ([]*IFace, error) {
	ifaces := []*IFace{}
	for _, pkg := range pkgs {
		for _, name := range pkg.Types.Scope().Names() {
			name, iface, err := loader.lookupIFace(pkg, name)
			if err == nil {
				// TODO: filter for matching name and file patterns
				ifaces = append(ifaces, NewIFace(name, pkg.Fset, iface))
			}
		}
	}
	return ifaces, nil
}
