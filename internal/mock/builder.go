package mock

import (
	"regexp"
	"strings"

	"github.com/tkrop/go-testing/internal/slices"
)

var (
	typeRegexp = regexp.MustCompile(`[a-zA-z0-9\./_-]*\.`)

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

// typeParam provides the parsed parameter type information.
type typeParam struct {
	param   *Param
	indexes [][]int
}

// FileBuilder provides the information
type FileBuilder struct {
	target  Type
	imports []*Import
	mocks   []*Mock

	paths   map[string]*Import
	aliases map[string]*Import
	params  []*typeParam
}

func NewFileBuilder(target Type) *FileBuilder {
	return &FileBuilder{
		target:  target,
		paths:   map[string]*Import{},
		aliases: map[string]*Import{},
	}
}

func (b *FileBuilder) AddMocks(mock ...*Mock) *FileBuilder {
	b.mocks = append(b.mocks, mock...)
	return b
}

func (b *FileBuilder) AddImports(imports ...*Import) *FileBuilder {
	b.imports = append(b.imports, imports...)

	for _, imprt := range imports {
		b.paths[imprt.Path] = imprt
		if _, ok := b.aliases[imprt.Alias]; ok {
			panic(NewErrAliasConflict(imprt.Path, imprt.Alias))
		}
		b.aliases[imprt.Alias] = imprt
	}

	return b
}

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

func (b *FileBuilder) paramImports(params []*Param) *FileBuilder {
	for _, param := range params {
		indexes := typeRegexp.FindAllStringIndex(param.Type, -1)
		if len(indexes) > 0 {
			b.params = append(b.params, &typeParam{
				param: param, indexes: indexes,
			})
			for _, index := range indexes {
				path := param.Type[index[0] : index[1]-1]
				if _, ok := b.paths[path]; !ok {
					b.paths[path] = nil
				}
			}
		}
	}

	return b
}

func (b *FileBuilder) buildImports() *FileBuilder {
	for path, imprt := range b.paths {
		if imprt != nil {
			continue
		}

		if strings.LastIndexAny(path, "/") < 0 {
			b.AddImports(&Import{Alias: path, Path: path})
			continue
		}

		b.AddImports(&Import{
			Alias: b.uniqAlias(path), Path: path,
		})
	}

	return b
}

func (b *FileBuilder) applyImports() *FileBuilder {
	for _, param := range b.params {
		builder := strings.Builder{}
		ptype, last := param.param.Type, 0
		for _, index := range param.indexes {
			builder.WriteString(ptype[last:index[0]])
			path := ptype[index[0] : index[1]-1]
			if imprt, ok := b.paths[path]; ok {
				builder.WriteString(imprt.Alias)
			}
			last = index[1] - 1
		}
		builder.WriteString(ptype[last:])
		param.param.Type = builder.String()
	}
	b.params = nil

	return b
}

func (b *FileBuilder) uniqAlias(path string) string {
	alias := ""

	norm := strings.ToLower(baseReplacer.Replace(path))
	for _, prefix := range slices.Reverse(strings.Split(norm, "/")) {
		if alias != "" {
			alias = prefix + "_" + alias
		} else {
			alias = prefix
		}

		if b.aliases[alias] == nil {
			return alias
		}
	}

	panic(NewErrAliasConflict(path, alias))
}
