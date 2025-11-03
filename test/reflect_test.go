package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type Struct struct {
	s string
	a any
}

func NewStruct(s string, a any) Struct     { return Struct{s: s, a: a} }
func NewPtrStruct(s string, a any) *Struct { return &Struct{s: s, a: a} }

var (
	structInit     = NewStruct("init", "init")
	structEmpty    = NewStruct("", nil)
	structFinal    = NewStruct("set final", "set final")
	structPtrInit  = NewPtrStruct("init", "init")
	structPtrEmpty = NewPtrStruct("", nil)
	structPtrFinal = NewPtrStruct("set final", "set final")
)

type BuilderStructParams struct {
	target Struct
	setup  func(test.Builder[Struct])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[Struct])
}

var builderStructTestCases = map[string]BuilderStructParams{
	"struct get init": {
		target: structInit,
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "init", b.Get("s"))
			assert.Equal(t, "init", b.Get("a"))
			assert.Equal(t, "init", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "init", b.Find("default"))
			assert.Equal(t, "init", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structInit, b.Get(""))
			assert.Equal(t, structInit, b.Build())
		},
	},

	"struct get invalid": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct set invalid": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct set compatible": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct set non-compatible": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"struct set": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct set nil": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set any").Set("a", "set any").
				Set("s", nil).Set("a", nil)
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct reset no pointer": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structFinal)
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[test_test.Struct => *test_test.Struct]"),
	},

	"struct reset pointer": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct reset any nil": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", nil)
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct reset struct nil": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", (*Struct)(nil))
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct reset any invalid": {
		target: structInit,
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *test_test.Struct]"),
	},
}

func TestBuilderStruct(t *testing.T) {
	test.Map(t, builderStructTestCases).
		Run(func(t test.Test, param BuilderStructParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := test.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// The
			param.check(t, accessor)
		})
}

type BuilderPtrStructParams struct {
	target *Struct
	setup  func(test.Builder[*Struct])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[*Struct])
}

var builderPtrStructTestCases = map[string]BuilderPtrStructParams{
	// Test cases for nil interface pointer.
	"nil any get": {
		target: nil,
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, "", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "", b.Find("default"))
			assert.Equal(t, "", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structPtrEmpty, b.Get(""))
			assert.Equal(t, structPtrEmpty, b.Build())
		},
	},

	"nil any get invalid": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil any set invalid": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil any set": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil any set compatible": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil any set non-compatible": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil any reset": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil any reset nil": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("", (*Struct)(nil))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, (*Struct)(nil), b.Get(""))
			assert.Equal(t, (*Struct)(nil), b.Build())
		},
	},

	"nil any reset invalid": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *test_test.Struct]"),
	},

	"nil any reset nil invalid": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("", (*Struct)(nil))
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	// Test cases for nil struct pointer.
	"nil struct get": {
		target: new(Struct),
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, "", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "", b.Find("default"))
			assert.Equal(t, "", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structPtrEmpty, b.Get(""))
			assert.Equal(t, structPtrEmpty, b.Build())
		},
	},

	"nil struct get invalid": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil struct set invalid": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil struct set": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil struct set compatible": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil struct set non-compatible": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil struct reset": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil struct reset invalid": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *test_test.Struct]"),
	},

	// Test cases for struct pointer instance.
	"ptr get": {
		target: NewPtrStruct("init", "init"),
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "init", b.Get("s"))
			assert.Equal(t, "init", b.Get("a"))
			assert.Equal(t, "init", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "init", b.Find("default"))
			assert.Equal(t, "init", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structPtrInit, b.Get(""))
			assert.Equal(t, structPtrInit, b.Build())
		},
	},

	"ptr get invalid": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr set": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr set compatible": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr set non-compatible": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"ptr reset": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr reset invalid": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *test_test.Struct]"),
	},
}

func TestBuilderPtrStruct(t *testing.T) {
	test.Map(t, builderPtrStructTestCases).
		Run(func(t test.Test, param BuilderPtrStructParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := test.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// Then
			param.check(t, accessor)
		})
}

type BuilderAnyParams struct {
	target any
	setup  func(test.Builder[any])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[any])
}

var builderAnyTestCases = map[string]BuilderAnyParams{
	// Test cases for invalid types.
	"invalid type nil": {
		target: nil,
		check: func(t test.Test, b test.Builder[any]) {
			assert.Nil(t, b)
		},
	},
	"invalid type int": {
		target: 1,
		check: func(t test.Test, b test.Builder[any]) {
			assert.Nil(t, b)
		},
	},

	// Test cases for struct instance.
	"struct get init": {
		target: structInit,
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "init", b.Get("s"))
			assert.Equal(t, "init", b.Get("a"))
			assert.Equal(t, "init", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "init", b.Find("default"))
			assert.Equal(t, "init", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structInit, b.Get(""))
			assert.Equal(t, structInit, b.Build())
		},
	},

	"struct get invalid": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct set invalid": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct set": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct set compatible": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct set non-compatible": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"struct reset pointer": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct reset no pointer": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structFinal)
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[test_test.Struct => *test_test.Struct]"),
	},

	"struct reset invalid": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *test_test.Struct]"),
	},

	// Test cases for struct pointer instance.
	"ptr get": {
		target: NewPtrStruct("init", "init"),
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "init", b.Get("s"))
			assert.Equal(t, "init", b.Get("a"))
			assert.Equal(t, "init", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "init", b.Find("default"))
			assert.Equal(t, "init", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structPtrInit, b.Get(""))
			assert.Equal(t, structPtrInit, b.Build())
		},
	},

	"ptr get invalid": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr set invalid": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr set": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr set compatible": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr set non-compatible": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"ptr reset": {
		target: NewPtrStruct("init", "init"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	// Test cases for nil struct pointer instance.
	"nil get": {
		target: new(Struct),
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, "", b.Find("default", "s"))
			assert.Equal(t, "default", b.Find("default", "a"))
			assert.Equal(t, "", b.Find("default"))
			assert.Equal(t, "", b.Find("default", "*"))
			assert.Equal(t, "default", b.Find("default", "x"))
			assert.Equal(t, structPtrEmpty, b.Get(""))
			assert.Equal(t, structPtrEmpty, b.Build())
		},
	},

	"nil get invalid": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil set invalid": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil set": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil set compatible": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil set non-compatible": {
		target: structInit,
		setup: func(b test.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil reset": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},
}

func TestBuilderAny(t *testing.T) {
	test.Map(t, builderAnyTestCases).
		Run(func(t test.Test, param BuilderAnyParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := test.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// The
			param.check(t, accessor)
		})
}

type FindParams struct {
	param  any
	deflt  any
	names  []string
	expect any
}

var findTestCases = map[string]FindParams{
	// Test cases for values.
	"int": {
		param:  1,
		deflt:  -1,
		expect: 1,
	},
	"bool": {
		param:  true,
		deflt:  false,
		expect: true,
	},
	"string": {
		param:  "init",
		deflt:  "default",
		expect: "init",
	},
	"invalid": {
		param:  "init",
		deflt:  true,
		expect: true,
	},

	"struct match": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"s"},
		expect: "init",
	},
	"struct invalid": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"invalid"},
		expect: "default",
	},
	"struct any": {
		param:  structInit,
		deflt:  "default",
		names:  []string{},
		expect: "init",
	},
	"struct star": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"invalid", "*"},
		expect: "init",
	},

	"ptr match": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"s"},
		expect: "init",
	},
	"ptr invalid": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"invalid"},
		expect: "default",
	},
	"ptr any": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{},
		expect: "init",
	},
	"ptr star": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"invalid", "*"},
		expect: "init",
	},
}

func TestFind(t *testing.T) {
	test.Map(t, findTestCases).
		Run(func(t test.Test, param FindParams) {
			// When
			expect := test.Find(param.param, param.deflt, param.names...)

			// Then
			assert.Equal(t, param.expect, expect)
		})
}

//revive:disable-next-line:function-length // Test suite approach.
func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("builder-struct", func(t *testing.T) {
		t.Parallel()
		// Given
		b := test.NewBuilder[Struct]()

		// When
		b.Set("s", "set final").Set("a", "set final")

		// Then
		assert.Equal(t, "set final", b.Get("s"))
		assert.Equal(t, "set final", b.Get("a"))
		assert.Equal(t, structFinal, b.Get(""))
		assert.Equal(t, structFinal, b.Build())
	})

	t.Run("builder-ptr", func(t *testing.T) {
		t.Parallel()
		// Given
		b := test.NewBuilder[*Struct]()

		// When
		b.Set("s", "set final").Set("a", "set final")

		// Then
		assert.Equal(t, "set final", b.Get("s"))
		assert.Equal(t, "set final", b.Get("a"))
		assert.Equal(t, structPtrFinal, b.Get(""))
		assert.Equal(t, structPtrFinal, b.Build())
	})

	t.Run("setter-nil", func(t *testing.T) {
		t.Parallel()

		// Given
		s := test.NewSetter((*Struct)(nil))

		// When
		s.Set("s", "set final").Set("a", "set final")

		// ThenstructEmpty
		assert.Equal(t, structPtrFinal, s.Build())
	})

	t.Run("setter-struct", func(t *testing.T) {
		t.Parallel()
		// Given
		s := test.NewSetter(NewStruct("init", "init"))

		// When
		s.Set("s", "set final").Set("a", "set final")

		// Then
		assert.Equal(t, structFinal, s.Build())
	})

	t.Run("setter-ptr", func(t *testing.T) {
		t.Parallel()

		// Given
		s := test.NewSetter(NewPtrStruct("init", "init"))

		// When
		s.Set("s", "set final").Set("a", "set final")

		// Then
		assert.Equal(t, structPtrFinal, s.Build())
	})

	t.Run("getter-nil", func(t *testing.T) {
		t.Parallel()

		// Given
		g := test.NewGetter((*Struct)(nil))

		// Then
		assert.Equal(t, "", g.Get("s"))
		assert.Equal(t, nil, g.Get("a"))
		assert.Equal(t, structPtrEmpty, g.Get(""))
	})

	t.Run("getter-struct", func(t *testing.T) {
		t.Parallel()

		// Given
		g := test.NewGetter(structFinal)

		// Then
		assert.Equal(t, "set final", g.Get("s"))
		assert.Equal(t, "set final", g.Get("a"))
		assert.Equal(t, structFinal, g.Get(""))
	})

	t.Run("getter-ptr", func(t *testing.T) {
		t.Parallel()

		// Given
		g := test.NewGetter(structPtrFinal)

		// Then
		assert.Equal(t, "set final", g.Get("s"))
		assert.Equal(t, "set final", g.Get("a"))
		assert.Equal(t, structPtrFinal, g.Get(""))
	})

	t.Run("finder-nil", func(t *testing.T) {
		t.Parallel()

		// Given
		f := test.NewFinder((*Struct)(nil))

		// Then
		assert.Equal(t, "", f.Find("default", "s"))
		assert.Equal(t, "default", f.Find("default", "a"))
		assert.Equal(t, "", f.Find("default"))
		assert.Equal(t, "", f.Find("default", "*"))
		assert.Equal(t, "default", f.Find("default", "x"))
	})

	t.Run("finder-struct", func(t *testing.T) {
		t.Parallel()

		// Given
		f := test.NewFinder(structFinal)

		// Then
		assert.Equal(t, "set final", f.Find("default", "s"))
		assert.Equal(t, "default", f.Find("default", "a"))
		assert.Equal(t, "set final", f.Find("default"))
		assert.Equal(t, "set final", f.Find("default", "*"))
		assert.Equal(t, "default", f.Find("default", "x"))
	})

	t.Run("finder-ptr", func(t *testing.T) {
		t.Parallel()

		// Given
		f := test.NewFinder(structPtrFinal)

		// Then
		assert.Equal(t, "set final", f.Find("default", "s"))
		assert.Equal(t, "default", f.Find("default", "a"))
		assert.Equal(t, "set final", f.Find("default"))
		assert.Equal(t, "set final", f.Find("default", "*"))
		assert.Equal(t, "default", f.Find("default", "x"))
	})
}
