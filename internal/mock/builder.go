package mock

import (
	"regexp"
	"strings"

	"github.com/tkrop/go-testing/internal/slices"
)

var (
	// Regexp to extract the qualifier information from type.
	qualifierRegexp = regexp.MustCompile(`[a-zA-z0-9\./_-]*\.`)
	// Replaceer to prepare the path to be used as alias qualifier.
	baseReplacer = strings.NewReplacer(
		"go-", "",
		"-go", "",
		"-", "/",
		".", "_",
		"@", "",
		"+", "",
		"~", "",
	)
)

// indexParam provides the parsed parameter type information.
type indexParam struct {
	param   *Param
	indexes [][]int
}

// FileBuilder is used to collect the mock file information used as input for
// the mock generator to create the mock source file using the template.
type FileBuilder struct {
	// Representation of file.
	target  *Type
	imports []*Import
	mocks   []*Mock

	// Maps for validating uniquness properties.
	paths   map[string]*Import
	aliases map[string]*Import

	// Slices to coordinate build steps.
	tpaths []string
	params []*indexParam
}

// NewFileBuilder creates a new file builder.
func NewFileBuilder(target *Type) *FileBuilder {
	return &FileBuilder{
		target:  target,
		paths:   map[string]*Import{},
		aliases: map[string]*Import{},
	}
}

// AddMocks adds a slice of mocks to the file.
func (b *FileBuilder) AddMocks(mock ...*Mock) *FileBuilder {
	b.mocks = append(b.mocks, mock...)
	return b
}

// AddImports adds a slice of imports to the file. If an import is matching
// the target import path, the alias is forcefully removed for this import.
func (b *FileBuilder) AddImports(imports ...*Import) *FileBuilder {
	for _, imprt := range imports {
		if imprt.Path == b.target.Path && imprt.Alias != "" {
			panic(NewErrIllegalImport(imprt))
		}
		b.addImport(imprt)
	}

	return b
}

// Builds the actual file information.
func (b *FileBuilder) Build() *File {
	for _, mock := range b.mocks {
		for _, method := range mock.Methods {
			b.paramImports(method.Params).
				paramImports(method.Results)
		}
	}

	b.buildImports().applyImports()

	return &File{
		Target:  b.target,
		Imports: b.imports,
		Mocks:   b.mocks,
	}
}

// addImport adds a single validated import to the paths and aliases slices.
// If the import has no alias, it is not added to the alias list, which is
// usually the case for the import of the target package.
func (b *FileBuilder) addImport(imprt *Import) {
	b.imports = append(b.imports, imprt)
	b.paths[imprt.Path] = imprt

	if imprt.Alias != "" {
		if conflict, ok := b.aliases[imprt.Alias]; ok {
			panic(NewErrAliasConflict(conflict, imprt.Path))
		}
		b.aliases[imprt.Alias] = imprt
	}
}

// paramImports collects the import path from the parameter types to prepare
// creating imports with minised alias names. For speedup and simplification
// the builder collects a list of parsed types with indexes to allow a quick
// exchange of the qualifier with the calculated alias.
func (b *FileBuilder) paramImports(params []*Param) *FileBuilder {
	for _, param := range params {
		indexes := qualifierRegexp.FindAllStringIndex(param.Type, -1)
		if len(indexes) > 0 {
			b.params = append(b.params, &indexParam{
				param: param, indexes: indexes,
			})
			for _, index := range indexes {
				path := param.Type[index[0] : index[1]-1]
				if _, ok := b.paths[path]; !ok {
					b.tpaths = append(b.tpaths, path)
					b.paths[path] = nil
				}
			}
		}
	}

	return b
}

// buildImports builds the imports for the collected paths to prepare applying
// the import alias to types of method argument and return parameters.
func (b *FileBuilder) buildImports() *FileBuilder {
	for _, path := range b.tpaths {
		if path == b.target.Path {
			b.addImport(&Import{Path: path})
			continue
		} else if strings.LastIndexAny(path, "/") < 0 {
			b.addImport(&Import{Alias: path, Path: path})
			continue
		}

		b.addImport(&Import{
			Alias: b.calcUniqAlias(path), Path: path,
		})
	}
	b.tpaths = nil

	return b
}

// applyImports applies the import alias to the types of method argument and
// return parameters replacing the package path.
func (b *FileBuilder) applyImports() *FileBuilder {
	for _, param := range b.params {
		builder := strings.Builder{}
		ptype, last := param.param.Type, 0
		for _, index := range param.indexes {
			builder.WriteString(ptype[last:index[0]])
			path := ptype[index[0] : index[1]-1]
			if imprt := b.paths[path]; imprt.Alias != "" {
				builder.WriteString(imprt.Alias)
				last = index[1] - 1
			} else {
				last = index[1]
			}
		}
		builder.WriteString(ptype[last:])
		param.param.Type = builder.String()
	}
	b.params = nil

	return b
}

// calcUniqAlias calculates a uniqu alias for the given package import path.
func (b *FileBuilder) calcUniqAlias(path string) string {
	alias := ""

	norm := strings.ToLower(baseReplacer.Replace(path))
	for _, prefix := range slices.Reverse(strings.Split(norm, "/")) {
		if alias != "" {
			alias = prefix + "_" + alias
		} else {
			alias = prefix
		}

		if _, ok := b.aliases[alias]; !ok {
			return alias
		}
	}

	panic(NewErrAliasConflict(&Import{Alias: alias, Path: path}, path))
}
