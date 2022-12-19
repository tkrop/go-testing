package test

import (
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testchan    = make(chan Test)
	testint     = 1
	testfloat   = 1.0
	testcomplex = complex(1.0, 1.0)
	teststring  = "value"
	testmap     = map[string]string{"value": "value"}
)

type GetValueParams struct {
	value  reflect.Value
	expect any
}

func FuncTest() {}

var testGetValueParams = map[string]GetValueParams{
	"bool": {
		value:  reflect.ValueOf(true),
		expect: true,
	},
	"int": {
		value:  reflect.ValueOf(int(1)),
		expect: int(1),
	},
	"int8": {
		value:  reflect.ValueOf(int8(1)),
		expect: int8(1),
	},
	"int16": {
		value:  reflect.ValueOf(int16(1)),
		expect: int16(1),
	},
	"int32": {
		value:  reflect.ValueOf(int32(1)),
		expect: int32(1),
	},
	"int64": {
		value:  reflect.ValueOf(int64(1)),
		expect: int64(1),
	},
	"intptr": {
		value:  reflect.ValueOf(&testint),
		expect: &testint,
	},
	"uint": {
		value:  reflect.ValueOf(uint(1)),
		expect: uint(1),
	},
	"uint8": {
		value:  reflect.ValueOf(uint8(1)),
		expect: uint8(1),
	},
	"uint16": {
		value:  reflect.ValueOf(uint16(1)),
		expect: uint16(1),
	},
	"uint32": {
		value:  reflect.ValueOf(uint32(1)),
		expect: uint32(1),
	},
	"uint64": {
		value:  reflect.ValueOf(uint64(1)),
		expect: uint64(1),
	},

	"float32": {
		value:  reflect.ValueOf(float32(1)),
		expect: float32(1),
	},
	"float64": {
		value:  reflect.ValueOf(float64(1)),
		expect: float64(1),
	},
	"floatptr": {
		value:  reflect.ValueOf(&testfloat),
		expect: &testfloat,
	},
	"complex64": {
		value:  reflect.ValueOf(complex64(testcomplex)),
		expect: complex64(testcomplex),
	},
	"complex128": {
		value:  reflect.ValueOf(testcomplex),
		expect: testcomplex,
	},
	"string": {
		value:  reflect.ValueOf("value"),
		expect: "value",
	},
	"stringptr": {
		value:  reflect.ValueOf(&teststring),
		expect: &teststring,
	},

	"array": {
		value:  reflect.ValueOf([1]string{"value"}),
		expect: [1]string{"value"},
	},
	"slice": {
		value:  reflect.ValueOf([]string{"value"}),
		expect: []string{"value"},
	},
	"map": {
		value:  reflect.ValueOf(testmap),
		expect: testmap,
	},
	"struct": {
		value:  reflect.ValueOf(GetValueParams{expect: "value"}),
		expect: GetValueParams{expect: "value"},
	},
	"chan": {
		value:  reflect.ValueOf(testchan),
		expect: testchan,
	},
	"func": {
		value:  reflect.ValueOf(FuncTest),
		expect: FuncTest,
	},
}

func TestGetValue(t *testing.T) {
	Map(t, testGetValueParams).Run(func(t Test, param GetValueParams) {
		// When
		value := getValue(param.value)

		// Then
		if param.value.Type().Kind() == reflect.Func {
			assert.NotNil(t, value)
			assert.True(t, reflect.TypeOf(value).Kind() == reflect.Func)
		} else {
			assert.Equal(t, param.expect, value)
		}
	})
}

type ExtractParams struct {
	name   string
	value  any
	deflt  any
	expect any
}

type BoolParams struct {
	value bool
}

type IntParams struct {
	value int
}

type StringParams struct {
	value string
}

type ExportParams struct {
	Value string
}

type StructParams struct {
	value BoolParams
}

var testExtractParams = map[string]ExtractParams{
	"no struct": {
		name:   "value",
		value:  "string",
		deflt:  true,
		expect: true,
	},

	"bool value": {
		name:   "any",
		value:  true,
		deflt:  false,
		expect: true,
	},

	"bool found": {
		name:   "value",
		value:  BoolParams{value: true},
		deflt:  false,
		expect: true,
	},

	"bool fallback": {
		name:   "fallback",
		value:  BoolParams{value: true},
		deflt:  true,
		expect: true,
	},

	"bool notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  true,
		expect: true,
	},

	"int value": {
		name:   "any",
		value:  2,
		deflt:  1,
		expect: 2,
	},

	"int found": {
		name:   "value",
		value:  IntParams{value: 2},
		deflt:  1,
		expect: 2,
	},

	"int fallback": {
		name:   "fallback",
		value:  IntParams{value: 2},
		deflt:  1,
		expect: 2,
	},

	"int notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  1,
		expect: 1,
	},

	"string value": {
		name:   "any",
		value:  "value",
		deflt:  "default",
		expect: "value",
	},

	"string found": {
		name:   "value",
		value:  StringParams{value: "value"},
		deflt:  "default",
		expect: "value",
	},

	"string fallback": {
		name:   "fallback",
		value:  StringParams{value: "fallback"},
		deflt:  "default",
		expect: "fallback",
	},

	"string notfound": {
		name:   "notfound",
		value:  BoolParams{},
		deflt:  "default",
		expect: "default",
	},

	"export found": {
		name:   "value",
		value:  ExportParams{Value: "value"},
		deflt:  "default",
		expect: "value",
	},

	"export fallback": {
		name:   "fallback",
		value:  ExportParams{Value: "fallback"},
		deflt:  "default",
		expect: "fallback",
	},

	"export notfound": {
		name:   "notfound",
		value:  BoolParams{},
		deflt:  "default",
		expect: "default",
	},

	"struct found": {
		name: "value",
		value: StructParams{
			value: BoolParams{value: true},
		},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: true},
	},

	"struct fallback": {
		name: "fallback",
		value: StructParams{
			value: BoolParams{value: true},
		},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: true},
	},

	"struct notfound": {
		name:   "notfound",
		value:  StringParams{},
		deflt:  BoolParams{value: false},
		expect: BoolParams{value: false},
	},
}

func TestExtract(t *testing.T) {
	Map(t, testExtractParams).Run(func(t Test, param ExtractParams) {
		// When
		value := extract(param.value, param.deflt, param.name)

		// Then
		assert.Equal(t, param.expect, value)
	})
}
