package mock

import (
	"fmt"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// Type is a generic type information.
type Type struct {
	// Path of the interface.
	Path string
	// File name of the interface.
	File string
	// Package of the interface.
	Package string
	// Name of the interface.
	Name string
}

// With extends a type information with the given type information in all
// places that are not empty.
func (t Type) With(etype Type) Type {
	if etype.Package != "" {
		t.Package = etype.Package
	}
	if etype.Path != "" {
		t.Path = etype.Path
	}
	if etype.File != "" {
		t.File = etype.File
	}
	if etype.Name != "" {
		t.Name = etype.Name
	}
	return t
}

// NewType creates a new type information based on the given named type.
func NewType(name *types.TypeName, fset *token.FileSet) Type {
	pos := fset.Position(name.Pos())
	return Type{
		Path:    name.Pkg().Path(),
		File:    fmt.Sprintf("%s:%d", filepath.Base(pos.Filename), pos.Line),
		Package: name.Pkg().Name(),
		Name:    name.Name(),
	}
}

// Param provides method parameterss.
type Param struct {
	// Method parameter.
	Name string
	// Method type.
	Type string
}

// NewParams creates a new parameter slice based on the given tuple providing
// the argument or return parameter list.
func NewParams(tuple *types.Tuple) []*Param {
	params := make([]*Param, 0, tuple.Len())
	for index := 0; index < tuple.Len(); index++ {
		param := tuple.At(index)
		params = append(params, &Param{
			Name: param.Name(),
			Type: ToAny(param.Type().String()),
		})
	}
	return params
}

// ToAny substitutes all `interfaces{}` types against the `any` type.
func ToAny(atype string) string {
	if strings.HasSuffix(atype, "interface{}") {
		return strings.TrimSuffix(atype, "interface{}") + "any"
	}
	return atype
}

// Method provides the method information.
type Method struct {
	// Name of method.
	Name string
	// Method arguments.
	Params []*Param
	// Method results.
	Results []*Param
	// Flag whether last argument is variadic.
	Variadic bool
}

// NewMethods creates a new method slice containing all methods of the given
// interface type.
func NewMethods(iface *types.Interface) []*Method {
	methods := make([]*Method, 0, iface.NumMethods())
	for index := 0; index < iface.NumMethods(); index++ {
		method := iface.Method(index)
		sign := method.Type().Underlying().(*types.Signature)
		methods = append(methods, &Method{
			Name:     method.Name(),
			Params:   NewParams(sign.Params()),
			Results:  NewParams(sign.Results()),
			Variadic: sign.Variadic(),
		})
	}
	return methods
}

// Interface information.
type IFace struct {
	// Source interface information.
	Source Type
	// Methods of source/target interface.
	Methods []*Method
}

// NewIFace creats a new interface information using given definition.
func NewIFace(
	name *types.TypeName, fset *token.FileSet, iface *types.Interface,
) *IFace {
	return &IFace{
		Source:  NewType(name, fset),
		Methods: NewMethods(iface),
	}
}

// Mock interface information.
type Mock struct {
	// Source interface information.
	Source Type
	// Target interface information.
	Target Type
	// Methods of source/target interface.
	Methods []*Method
}

// Import information.
type Import struct {
	// Alias name of import.
	Alias string
	// Path name of full import.
	Path string
}

// File information.
type File struct {
	// Target file information.
	Target Type
	// Common import data.
	Imports []*Import
	// Mock interface data.
	Mocks []*Mock
	// Writer used for file output.
	Writer *os.File
}

// NewFiles creates a list of new mock files from the given set of mocks and
// given set of default imports to consider for applying for the template using
// the mocks target information to collect all mocks that should be written to
// the same file.
func NewFiles(mocks []*Mock, imports ...*Import) []*File {
	builders := []*FileBuilder{}

	bmap := map[Type]*FileBuilder{}
	for _, mock := range mocks {
		target := mock.Target
		target.Name = "" // file target must ignore target name!
		if builder, ok := bmap[target]; !ok {
			builder := NewFileBuilder(target).AddMocks(mock)
			builders = append(builders, builder)
			bmap[target] = builder
		} else {
			builder.AddMocks(mock)
		}
	}

	files := []*File{}
	for _, builder := range builders {
		file := builder.AddImports(imports...).Build()
		files = append(files, file)
	}

	return files
}

// Open opens a file descriptor for writing.
func (file *File) Open(stdout *os.File) error {
	target := &file.Target
	if target.File == "-" {
		file.Writer = stdout
		return nil
	}

	stdout, err := os.OpenFile(target.File,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o0600)
	file.Writer = stdout
	if err != nil {
		return NewErrFileOpening(target.File, err)
	}
	return nil
}

// Write writes the mocks using the given template to the file target.
func (file *File) Write(temp Template) error {
	err := temp.Execute(file.Writer, file)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the file descriptor for writing.
func (file *File) Close() error {
	target := &file.Target
	if target.File == "-" {
		return nil
	}

	err := file.Writer.Close()
	if err != nil {
		return err
	}
	return nil
}
