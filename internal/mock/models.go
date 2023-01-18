package mock

import (
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
)

// Source is the parsers source information.
type Source struct {
	// Path of the source package (files).
	Path string
	// Name of the source interface.
	Name string
}

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
}
