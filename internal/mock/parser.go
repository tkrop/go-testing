package mock

import (
	"go/token"
	"strings"
)

const (
	// Default directory for package loading.
	DirDefault = "."
	// Magic constant defining packages read from file.
	ReadFromFile = "command-line-arguments"
	// Default value of mock interface name pattern.
	MockPatternDefault = "Mock*"
	// Default value of interface regex matching pattern.
	MatchPatternDefault = ".*"
)

// argType defines the command line argument types.
type argType int

// Collection of command line argument types.
const (
	// Unknown argument type that results in an error.
	argTypeUnknown argType = iota
	// Invalid argument type that results in an not found error.
	argTypeNotFound

	// Target package argument type. Must be an identifier.
	argTypeSourcePkg
	// Source import path argument type. Must resolve to a loadable package.
	argTypeSourcePath
	// Source file or local path argument type. Must resolve to a loadable file
	// or a package.
	argTypeSourceFile

	// Target package argument type. Must be an identifier.
	argTypeTargetPkg
	// Target import path argument type. Must resolve to a loadable package.
	argTypeTargetPath
	// Target file or path argument type. Must be a relative or absolute path
	// containing a file. Auto detection requires and ending with `.go`.
	argTypeTargetFile

	// Source/target interface mapping argument type. Must be a list of
	// identifier mappings.
	argTypeIFace
)

// Parser is a mock setup command line argument parser.
type Parser struct {
	// A caching, reusable package loader.
	loader Loader
	// The default target definition for parsing.
	target *Type
}

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
	source *Type
	// Target provides the actual target interface setup.
	target *Type
}

// NewParser creates a new mock setup command line argument parser with loading
// of test packages enabled or disabled.
func NewParser(loader Loader, target *Type) *Parser {
	return &Parser{loader: loader, target: target}
}

// parse parses the command line arguments provided and returns a list of mock
// stubs containing the necessary source and target information without method
// data.
func (parser *Parser) Parse(args ...string) ([]*Mock, []error) {
	state := newParseState(parser)
	for pos, rarg := range args {
		switch atype, arg := parser.argType(rarg); atype {
		case argTypeSourcePkg:
			state.ensureIFace(pos - 1)
			state.source.Package = arg
		case argTypeSourcePath:
			state.ensureIFace(pos - 1)
			state.source = &Type{Path: arg}
		case argTypeSourceFile:
			state.ensureIFace(pos - 1)
			state.source = &Type{File: arg}
		case argTypeTargetPkg:
			state.target.Package = arg
		case argTypeTargetPath:
			state.target.Path = arg
		case argTypeTargetFile:
			state.target.File = arg
		case argTypeIFace:
			for _, arg := range strings.Split(arg, ",") {
				state.creatMocks(pos, arg)
			}
		case argTypeNotFound:
			err := NewErrArgNotFound(pos, arg)
			state.errs = append(state.errs, err)
		case argTypeUnknown:
			fallthrough
		default:
			err := NewErrArgInvalid(pos, arg)
			state.errs = append(state.errs, err)
		}
	}

	state.ensureSource().ensureIFace(len(args) - 1)
	if len(state.errs) != 0 {
		return nil, state.errs
	}
	return state.mocks, nil
}

// argType evaluates the a command line argument type by analysing argument
// value and returns the argument type with the remaining argument value.
func (parser *Parser) argType(arg string) (argType, string) {
	if strings.Index(arg, "--") == 0 {
		return parser.argTypeParse(arg)
	}
	return parser.argTypeGuess(arg)
}

// argTypeParse parses the type from the command line argument by looking at
// the flag and returing the remaining argument value. If the flag is open for
// ambiguaty, the concreat type is evaluated checking argument patterns and
// trying to load related packages.
func (parser *Parser) argTypeParse(arg string) (argType, string) {
	equal := strings.Index(arg, "=")
	flag, sarg := arg[2:equal], arg[equal+1:]
	switch flag {
	case "source":
		return parser.argTypeSource(sarg)
	case "source-pkg":
		return argTypeSourcePkg, sarg
	case "source-path":
		return argTypeSourcePath, sarg
	case "source-file":
		return argTypeSourceFile, sarg

	case "target":
		return parser.argTypeTarget(sarg)
	case "target-pkg":
		return argTypeTargetPkg, sarg
	case "target-path":
		return argTypeTargetPath, sarg
	case "target-file":
		return argTypeTargetFile, sarg

	case "iface":
		return argTypeIFace, sarg
	default:
		return argTypeUnknown, arg
	}
}

// argTypeSource evaluates the concreate source argument type by trying to load
// the related package.
func (parser *Parser) argTypeSource(sarg string) (argType, string) {
	if token.IsIdentifier(sarg) {
		return argTypeSourcePkg, sarg
	}
	pkgs, err := parser.loader.Load(sarg).Get()
	if len(pkgs) > 0 && err == nil {
		if pkgs[0].PkgPath == ReadFromFile {
			return argTypeSourceFile, sarg
		}
		return argTypeSourcePath, sarg
	}
	return argTypeNotFound, sarg
}

// argTypeTarget evaluates the actual target argument type.
func (parser *Parser) argTypeTarget(sarg string) (argType, string) {
	if token.IsIdentifier(sarg) {
		return argTypeTargetPkg, sarg
	}
	pkgs, err := parser.loader.Load(sarg).Get()
	if len(pkgs) > 0 && err == nil {
		if pkgs[0].PkgPath == ReadFromFile {
			return argTypeTargetFile, sarg
		}
		return argTypeTargetPath, sarg
	}
	return argTypeTargetFile, sarg
}

// argTypeGuess evaluates the argument type by analysing the argument values
// and guessing their meaning by checking argument patterns and trying to load
// related packages.
func (parser *Parser) argTypeGuess(arg string) (argType, string) {
	if strings.ContainsAny(arg, "=,") {
		return argTypeIFace, arg
	} else if token.IsIdentifier(arg) {
		if token.IsExported(arg) {
			return argTypeIFace, arg
		}
		return argTypeTargetPkg, arg
	}

	if strings.HasSuffix(arg, ".go") {
		file := arg
		if index := strings.LastIndex(arg, "/"); index >= 0 {
			file = arg[index+1:]
		}
		if strings.HasPrefix(file, "mock_") {
			return argTypeTargetFile, arg
		}
	}

	pkgs, err := parser.loader.Load(arg).Get()
	if len(pkgs) > 0 && err == nil {
		if pkgs[0].PkgPath == ReadFromFile {
			return argTypeSourceFile, arg
		}
		return argTypeSourcePath, arg
	}
	return argTypeNotFound, arg
}

// newParseState creates a new parse state for parsing.
func newParseState(parser *Parser) *ParseState {
	return &ParseState{
		loader:  parser.loader,
		source:  &Type{},
		target:  parser.target.Copy(),
		targets: map[Type]*Mock{},
		mocks:   []*Mock{},
		errs:    []error{},
	}
}

// ensureSource checks the current source setup and derives the default package
// name and package path from the given source file, if the setup is missing.
// The default name is ensured when parsing the interface specific command line
// arguments.
func (state *ParseState) ensureSource() *ParseState {
	source := state.source
	if source.File == "" && source.Path == "" {
		source.File = DirDefault
	}
	return state
}

// ensureIFace ensures that the last source setup was used to create at least
// one mock stub before we continue with the next source package setup. If no
// mock stub was created, a generic mock stub is created that mocks all
// interfaces in the source package.
func (state *ParseState) ensureIFace(pos int) {
	if state.source.IsPartial() {
		state.creatMocks(pos, "")
	}
}

// ensureState ensures that the source and target are setup with package names,
// package paths, and source interface and target mock name patthers from the
// provided interface command line argument. If the command line argument is
// incomplete the names are setup with the default source and target mock names.
func (state *ParseState) ensureState(arg string) {
	state.source.Update(state.loader)
	state.target.Update(state.loader)

	if arg == "" {
		state.source.Name = MatchPatternDefault
		state.target.Name = MockPatternDefault
		return
	}

	index := strings.IndexAny(arg, "=")
	switch index {
	case -1:
		state.source.Name = arg
		state.target.Name = MockPatternDefault
	case 0:
		state.source.Name = MatchPatternDefault
		state.target.Name = arg[1:]
	case len(arg) - 1:
		state.source.Name = arg[:index]
		state.target.Name = MockPatternDefault
	default:
		state.source.Name = arg[:index]
		state.target.Name = arg[index+1:]
	}
}

// creatMocks creates the mock stubs for the given source/target interface
// command line argument. If no target interface name/pattern is provided it
// is set to the source interface name prefixed by `Mock`. If no source
// interface name/pattern is provided it is assumed to be any interface in the
// requested package.
func (state *ParseState) creatMocks(pos int, arg string) {
	state.ensureState(arg)

	ifaces, err := state.loader.IFaces(state.source)
	if err != nil {
		state.errs = append(state.errs, NewErrArgFailure(pos, arg, err))
	}

	for _, iface := range ifaces {
		target := *state.target
		target.Name = strings.ReplaceAll(target.Name, "*", iface.Source.Name)
		if _, ok := state.targets[target]; !ok {
			mock := &Mock{
				Source:  iface.Source,
				Target:  &target,
				Methods: iface.Methods,
			}
			state.targets[target] = mock
			state.mocks = append(state.mocks, mock)
		}
	}
}
