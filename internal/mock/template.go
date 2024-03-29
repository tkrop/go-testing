package mock

import (
	"io"
	"strconv"
	"strings"
	"text/template"
)

var (
	ImportReflect = &Import{
		Alias: "reflect", Path: "reflect",
	}
	ImportGomock = &Import{
		Alias: "gomock", Path: "github.com/golang/mock/gomock",
	}
	ImportMock = &Import{
		Alias: "mock", Path: "github.com/tkrop/go-testing/mock",
	}
	ImportsTemplate = []*Import{
		ImportReflect, ImportGomock, ImportMock,
	}

	MockFileFuncMap = template.FuncMap{
		"ImportsList": importArgs,
		"ParamArgs":   paramArgs,
		"ResultArgs":  resultArgs,
		"CallArgs":    callArgs,
		"ConvertArgs": convertArgs,
	}

	MockFileTemplate = `// Code generated by mock; DO NOT EDIT.

// Package {{.Target.Package}} or better this file is auto generated by
// github.com/tkrop/go-testing/cmd/mock.
package {{.Target.Package}}

import (
{{.Imports | ImportsList -}}
)
{{- range $index, $mock := .Mocks}}

// {{$mock.Target.Name}} is a mock of {{$mock.Source.Name}}.
//
// Source: {{$mock.Source.Path}}/{{$mock.Source.File}}.
type {{$mock.Target.Name}} struct {
	ctrl     *gomock.Controller
	recorder *{{$mock.Target.Name}}Recorder
}

// {{$mock.Target.Name}}Recorder is the mock recorder for {{$mock.Target.Name}}.
type {{$mock.Target.Name}}Recorder struct {
	mock *{{$mock.Target.Name}}
}

// New{{$mock.Target.Name}} creates a new mock instance.
func New{{$mock.Target.Name}}(ctrl *gomock.Controller) *{{$mock.Target.Name}} {
	mock := &{{$mock.Target.Name}}{ctrl: ctrl}
	mock.recorder = &{{$mock.Target.Name}}Recorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *{{$mock.Target.Name}}) EXPECT() *{{$mock.Target.Name}}Recorder {
	return m.recorder
}

{{- if not .Methods}}

// {{$mock.Target.Name}} has no methods.
{{- end -}}

{{- range $index, $method := .Methods}}

// {{$method.Name}} is the mock method to capture a coresponding call.
func (m *{{$mock.Target.Name}}) {{$method.Name}}(
	{{- if $method.Params -}}
	{{$method.Params | ParamArgs}}
	{{- end -}}
)
{{- if $method.Results}} {{$method.Results | ResultArgs }}
{{- end }} {
	m.ctrl.T.Helper()
	{{if $method.Results}}ret := {{ end -}} m.ctrl.Call(m, "{{$method.Name}}"
		{{- if $method.Params}}, {{$method.Params | CallArgs}}{{- end}})

	{{- if $method.Results}}

	{{$method.Results | ConvertArgs}}
	{{- end}}

	{{- if $method.Results}}

	return {{$method.Results | CallArgs}}
	{{- end}}
}

// {{$method.Name}} is the recorder method to indicates an expected call.
func (mr *{{$mock.Target.Name}}Recorder) {{$method.Name}}(
	{{- if $method.Params -}}
	{{$method.Params | ParamArgs}}
	{{- end -}}
) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "{{$method.Name}}",
		reflect.TypeOf((*{{$mock.Target.Name}})(nil).{{$method.Name}})
		{{- if $method.Params}}, {{$method.Params | CallArgs}} {{- end}})
}
{{- end -}}
{{- end}}
`
)

// Template minimal mock template abstraction.
type Template interface {
	Execute(writer io.Writer, file any) error
}

// NewTemplate creates a new mock template.
func NewTemplate() (Template, []*Import, error) {
	temp, err := template.New("file").
		Funcs(MockFileFuncMap).Parse(MockFileTemplate)
	return temp, ImportsTemplate, err //nolint:wrapcheck  // needs consideration
}

// MustTemplate crates a new mock template panicing in case of errors.
func MustTemplate() Template {
	return template.Must(template.New("file").
		Funcs(MockFileFuncMap).Parse(MockFileTemplate))
}

func importArgs(imports []*Import) string {
	// TODO: order imports compliant with linters.

	builder := strings.Builder{}
	for _, imprt := range imports {
		if imprt.Alias == "" {
			continue
		}
		builder.WriteRune('\t')
		if !strings.HasSuffix(imprt.Path, imprt.Alias) {
			builder.WriteString(imprt.Alias)
			builder.WriteRune(' ')
		}
		builder.WriteRune('"')
		builder.WriteString(imprt.Path)
		builder.WriteString("\"\n")
	}
	return builder.String()
}

func paramArgs(params []*Param) string {
	builder := strings.Builder{}
	mindex := len(params) - 1
	for index, param := range params {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(param.Name)
		if index == mindex || params[index+1].Type != param.Type {
			builder.WriteRune(' ')
			builder.WriteString(param.Type)
		}
	}
	return builder.String()
}

func resultArgs(params []*Param) string {
	builder := strings.Builder{}
	if len(params) > 1 {
		builder.WriteRune('(')
	}
	for index, result := range params {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(result.Type)
	}
	if len(params) > 1 {
		builder.WriteRune(')')
	}
	return builder.String()
}

func callArgs(params []*Param) string {
	builder := strings.Builder{}
	for index, param := range params {
		if index > 0 {
			builder.WriteString(", ")
		}
		if param.Name != "" {
			builder.WriteString(param.Name)
		} else {
			builder.WriteString("ret")
			builder.WriteString(strconv.Itoa(index))
		}
	}
	return builder.String()
}

func convertArgs(params []*Param) string {
	builder := strings.Builder{}
	for index, param := range params {
		istr := strconv.Itoa(index)
		if index > 0 {
			builder.WriteString("\n\t")
		}
		if param.Name != "" {
			builder.WriteString(param.Name)
		} else {
			builder.WriteString("ret")
			builder.WriteString(istr)
		}
		builder.WriteString(", _ := ret[")
		builder.WriteString(istr)
		builder.WriteString("].(")
		builder.WriteString(param.Type)
		builder.WriteString(")")
	}
	return builder.String()
}
