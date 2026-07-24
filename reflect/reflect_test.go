package reflect_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/reflect"
	"github.com/tkrop/go-testing/test"
)

// Values used in the tests.
var (
	structInit     = newStruct("init", "init")
	structEmpty    = newStruct("", nil)
	structFinal    = newStruct("set final", "set final")
	structPtrInit  = newPtrStruct("init", "init")
	structPtrEmpty = newPtrStruct("", nil)
	structPtrFinal = newPtrStruct("set final", "set final")
)

// Types used in the tests.
type (
	// intAlias is a test type alias for int.
	intAlias int
	// structPtrAlias is a test type alias for *Struct.
	structPtrAlias *structAny
	// structAny is a test struct type used for testing the `Builder` interface
	// with struct targets.
	structAny struct {
		s string
		a any
	}
)

// newStruct creates a new instance of `Struct` with the given string and any
// value.
func newStruct(s string, a any) structAny { return structAny{s: s, a: a} }

// newPtrStruct creates a new instance of `*Struct` with the given string and
// any value.
func newPtrStruct(s string, a any) *structAny { return &structAny{s: s, a: a} }

// builderStructParams is a test parameter type for testing the `Builder`
// interface with struct targets.
type builderStructParams struct {
	target structAny
	setup  func(reflect.Builder[structAny])
	expect mock.SetupFunc
	check  func(test.Test, reflect.Builder[structAny])
}

// Test cases for testing the `Builder` interface with struct targets.
var builderStructTestCases = map[string]builderStructParams{
	"struct-get-init": {
		target: structInit,
		check: func(t test.Test, b reflect.Builder[structAny]) {
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

	"struct-get-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct-set-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct-set-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-set-non-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"struct-set": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-set-nil": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set any").Set("a", "set any").
				Set("s", nil).Set("a", nil)
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct-reset-no-pointer": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structFinal)
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[reflect_test.structAny => *reflect_test.structAny]"),
	},

	"struct-reset-pointer": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-reset-any-nil": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", nil)
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct-reset-struct-nil": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", (*structAny)(nil))
		},
		check: func(t test.Test, b reflect.Builder[structAny]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, nil, b.Get("a"))
			assert.Equal(t, structEmpty, b.Get(""))
			assert.Equal(t, structEmpty, b.Build())
		},
	},

	"struct-reset-any-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *reflect_test.structAny]"),
	},
}

// TestBuilderStruct tests the `Builder` interface with struct targets using
// various test cases defined in the `builderStructTestCases` map.
func TestBuilderStruct(t *testing.T) {
	test.Map(t, builderStructTestCases).
		Run(func(t test.Test, param builderStructParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := reflect.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// The
			param.check(t, accessor)
		})
}

// BuilderAnyParams is a test parameter type for testing the `Builder`
// interface with any type of target.
type builderPtrStructParams struct {
	target *structAny
	setup  func(reflect.Builder[*structAny])
	expect mock.SetupFunc
	check  func(test.Test, reflect.Builder[*structAny])
}

// Test cases for testing the `Builder` interface with struct pointer targets.
var builderPtrStructTestCases = map[string]builderPtrStructParams{
	// Test cases for nil interface pointer.
	"nil-any-get": {
		target: nil,
		check: func(t test.Test, b reflect.Builder[*structAny]) {
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

	"nil-any-get-invalid": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-any-set-invalid": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-any-set": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-any-set-compatible": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-any-set-non-compatible": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil-any-reset": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-any-reset-nil": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("", (*structAny)(nil))
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, (*structAny)(nil), b.Get(""))
			assert.Equal(t, (*structAny)(nil), b.Build())
		},
	},

	"nil-any-reset-invalid": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *reflect_test.structAny]"),
	},

	"nil-any-reset-nil-invalid": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("", (*structAny)(nil))
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	// Test cases for nil struct pointer.
	"nil-struct-get": {
		target: new(structAny),
		check: func(t test.Test, b reflect.Builder[*structAny]) {
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

	"nil-struct-get-invalid": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-struct-set-invalid": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-struct-set": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-struct-set-compatible": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-struct-set-non-compatible": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil-struct-reset": {
		target: new(structAny),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-struct-reset-invalid": {
		target: nil,
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *reflect_test.structAny]"),
	},

	// Test cases for struct pointer instance.
	"ptr-get": {
		target: newPtrStruct("init", "init"),
		check: func(t test.Test, b reflect.Builder[*structAny]) {
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

	"ptr-get-invalid": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr-set": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr-set-compatible": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr-set-non-compatible": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"ptr-reset": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[*structAny]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr-reset-invalid": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[*structAny]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *reflect_test.structAny]"),
	},
}

// TestBuilderPtrStruct tests the `Builder` interface with struct pointer
// targets using various test cases defined in the `builderPtrStructTestCases`
// map.
func TestBuilderPtrStruct(t *testing.T) {
	test.Map(t, builderPtrStructTestCases).
		Run(func(t test.Test, param builderPtrStructParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := reflect.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// Then
			param.check(t, accessor)
		})
}

// builderAnyParams is a test parameter type for testing the Builder interface
// with any type of target.
type builderAnyParams struct {
	target any
	setup  func(reflect.Builder[any])
	expect mock.SetupFunc
	check  func(test.Test, reflect.Builder[any])
}

// Test cases for testing the Builder interface with any type of target.
var builderAnyTestCases = map[string]builderAnyParams{
	// Test cases for invalid types.
	"invalid-type-nil": {
		target: nil,
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Nil(t, b)
		},
	},
	"invalid-type-int": {
		target: 1,
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Nil(t, b)
		},
	},

	// Test cases for struct instance.
	"struct-get-init": {
		target: structInit,
		check: func(t test.Test, b reflect.Builder[any]) {
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

	"struct-get-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct-set-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"struct-set": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-set-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-set-non-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"struct-reset-pointer": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structFinal, b.Get(""))
			assert.Equal(t, structFinal, b.Build())
		},
	},

	"struct-reset-no-pointer": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structFinal)
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[reflect_test.structAny => *reflect_test.structAny]"),
	},

	"struct-reset-invalid": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", struct{}{})
		},
		expect: test.Panic("target must be compatible struct pointer " +
			"[struct {} => *reflect_test.structAny]"),
	},

	// Test cases for struct pointer instance.
	"ptr-get": {
		target: newPtrStruct("init", "init"),
		check: func(t test.Test, b reflect.Builder[any]) {
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

	"ptr-get-invalid": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr-set-invalid": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"ptr-set": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr-set-compatible": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"ptr-set-non-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"ptr-reset": {
		target: newPtrStruct("init", "init"),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	// Test cases for nil struct pointer instance.
	"nil-get": {
		target: new(structAny),
		check: func(t test.Test, b reflect.Builder[any]) {
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

	"nil-get-invalid": {
		target: new(structAny),
		setup: func(b reflect.Builder[any]) {
			b.Get("invalid")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-set-invalid": {
		target: new(structAny),
		setup: func(b reflect.Builder[any]) {
			b.Set("invalid", "set final")
		},
		expect: test.Panic("target field not found [invalid]"),
	},

	"nil-set": {
		target: new(structAny),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("s", "set final").Set("a", "set final")
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-set-compatible": {
		target: new(structAny),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", string([]byte("set final"))).
				Set("a", string([]byte("set final")))
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},

	"nil-set-non-compatible": {
		target: structInit,
		setup: func(b reflect.Builder[any]) {
			b.Set("s", []byte("set final"))
		},
		expect: test.Panic("value must be compatible [[]uint8 => string]"),
	},

	"nil-reset": {
		target: new(structAny),
		setup: func(b reflect.Builder[any]) {
			b.Set("s", "set first").Set("a", "set first").
				Set("", structPtrFinal)
		},
		check: func(t test.Test, b reflect.Builder[any]) {
			assert.Equal(t, "set final", b.Get("s"))
			assert.Equal(t, "set final", b.Get("a"))
			assert.Equal(t, structPtrFinal, b.Get(""))
			assert.Equal(t, structPtrFinal, b.Build())
		},
	},
}

// TestBuilderAny tests the Builder interface with any type of target.
func TestBuilderAny(t *testing.T) {
	test.Map(t, builderAnyTestCases).
		Run(func(t test.Test, param builderAnyParams) {
			// Given
			mock.NewMocks(t).Expect(param.expect)
			accessor := reflect.NewAccessor(param.target)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// Then
			param.check(t, accessor)
		})
}

// newBuilderParams is a test parameter type for testing the constructor
// functions of the reflect package.
type newBuilderParams struct {
	setup mock.SetupFunc
	call  func() any
	check func(t test.Test, b any)
}

// newBuilderTestCases is a map of test parameters for testing the constructor
// functions of the reflect package.
var newBuilderTestCases = map[string]newBuilderParams{
	"builder-ptr-alias-panic": {
		setup: test.Panic(fmt.Sprintf(
			"cast failed [%T]: %v", structPtrAlias(nil),
			newPtrStruct("", nil))),
		call: func() any {
			return reflect.NewBuilder[structPtrAlias]()
		},
	},

	"builder-struct": {
		call: func() any {
			b := reflect.NewBuilder[structAny]()
			b.Set("s", "set final").Set("a", "set final")

			return b
		},
		check: func(t test.Test, b any) {
			builder := test.Cast[reflect.Builder[structAny]](b)

			assert.Equal(t, "set final", builder.Get("s"))
			assert.Equal(t, "set final", builder.Get("a"))
			assert.Equal(t, structFinal, builder.Get(""))
			assert.Equal(t, structFinal, builder.Build())
		},
	},

	"builder-ptr": {
		call: func() any {
			b := reflect.NewBuilder[*structAny]()
			b.Set("s", "set final").Set("a", "set final")

			return b
		},
		check: func(t test.Test, b any) {
			builder := test.Cast[reflect.Builder[*structAny]](b)

			assert.Equal(t, "set final", builder.Get("s"))
			assert.Equal(t, "set final", builder.Get("a"))
			assert.Equal(t, structPtrFinal, builder.Get(""))
			assert.Equal(t, structPtrFinal, builder.Build())
		},
	},

	"setter-nil": {
		call: func() any {
			s := reflect.NewSetter((*structAny)(nil))
			s.Set("s", "set final").Set("a", "set final")

			return s
		},
		check: func(t test.Test, b any) {
			s := test.Cast[reflect.Setter[*structAny]](b)

			assert.Equal(t, structPtrFinal, s.Build())
		},
	},

	"setter-struct": {
		call: func() any {
			s := reflect.NewSetter(newStruct("init", "init"))
			s.Set("s", "set final").Set("a", "set final")

			return s
		},
		check: func(t test.Test, b any) {
			s := test.Cast[reflect.Setter[structAny]](b)

			assert.Equal(t, structFinal, s.Build())
		},
	},

	"setter-ptr": {
		call: func() any {
			s := reflect.NewSetter(newPtrStruct("init", "init"))
			s.Set("s", "set final").Set("a", "set final")

			return s
		},
		check: func(t test.Test, b any) {
			s := test.Cast[reflect.Setter[*structAny]](b)

			assert.Equal(t, structPtrFinal, s.Build())
		},
	},

	"getter-nil": {
		call: func() any {
			return reflect.NewGetter((*structAny)(nil))
		},
		check: func(t test.Test, b any) {
			g := test.Cast[reflect.Getter[*structAny]](b)

			assert.Equal(t, "", g.Get("s"))
			assert.Equal(t, nil, g.Get("a"))
			assert.Equal(t, structPtrEmpty, g.Get(""))
		},
	},

	"getter-struct": {
		call: func() any {
			return reflect.NewGetter(structFinal)
		},
		check: func(t test.Test, b any) {
			g := test.Cast[reflect.Getter[structAny]](b)

			assert.Equal(t, "set final", g.Get("s"))
			assert.Equal(t, "set final", g.Get("a"))
			assert.Equal(t, structFinal, g.Get(""))
		},
	},

	"getter-ptr": {
		call: func() any {
			return reflect.NewGetter(structPtrFinal)
		},
		check: func(t test.Test, b any) {
			g := test.Cast[reflect.Getter[*structAny]](b)

			assert.Equal(t, "set final", g.Get("s"))
			assert.Equal(t, "set final", g.Get("a"))
			assert.Equal(t, structPtrFinal, g.Get(""))
		},
	},

	"finder-nil": {
		call: func() any {
			return reflect.NewFinder((*structAny)(nil))
		},
		check: func(t test.Test, b any) {
			f := test.Cast[reflect.Finder[*structAny]](b)

			assert.Equal(t, "", f.Find("default", "s"))
			assert.Equal(t, "default", f.Find("default", "a"))
			assert.Equal(t, "", f.Find("default"))
			assert.Equal(t, "", f.Find("default", "*"))
			assert.Equal(t, "default", f.Find("default", "x"))
		},
	},

	"finder-struct": {
		call: func() any {
			return reflect.NewFinder(structFinal)
		},
		check: func(t test.Test, b any) {
			f := test.Cast[reflect.Finder[structAny]](b)

			assert.Equal(t, "set final", f.Find("default", "s"))
			assert.Equal(t, "default", f.Find("default", "a"))
			assert.Equal(t, "set final", f.Find("default"))
			assert.Equal(t, "set final", f.Find("default", "*"))
			assert.Equal(t, "default", f.Find("default", "x"))
		},
	},

	"finder-ptr": {
		call: func() any {
			return reflect.NewFinder(structPtrFinal)
		},
		check: func(t test.Test, b any) {
			f := test.Cast[reflect.Finder[*structAny]](b)

			assert.Equal(t, "set final", f.Find("default", "s"))
			assert.Equal(t, "default", f.Find("default", "a"))
			assert.Equal(t, "set final", f.Find("default"))
			assert.Equal(t, "set final", f.Find("default", "*"))
			assert.Equal(t, "default", f.Find("default", "x"))
		},
	},
}

// TestNewBuilder tests the constructor functions of the reflect package using
// various test cases defined in the `newBuilderTestCases` map.
func TestNewBuilder(t *testing.T) {
	test.Map(t, newBuilderTestCases).
		Run(func(t test.Test, param newBuilderParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			b := param.call()

			// Then
			if param.check != nil {
				param.check(t, b)
			}
		})
}

// findParams is a test parameter type for testing the `Find` function of the
// reflect package.
type findParams struct {
	setup  mock.SetupFunc
	param  any
	deflt  any
	names  []string
	call   func(findParams) any
	expect any
}

// findCall is a default call function for testing the `Find` function of the
// reflect package when a custom call function is not provided in the test
// parameters.
var findCall = func(param findParams) any {
	return reflect.Find(param.param, param.deflt, param.names...)
}

// findTestCases is a map of test parameters for testing the `Find` function
// of the reflect package.
var findTestCases = map[string]findParams{
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

	"struct-match": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"s"},
		expect: "init",
	},
	"struct-invalid": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"invalid"},
		expect: "default",
	},
	"struct-any": {
		param:  structInit,
		deflt:  "default",
		names:  []string{},
		expect: "init",
	},
	"struct-star": {
		param:  structInit,
		deflt:  "default",
		names:  []string{"invalid", "*"},
		expect: "init",
	},

	"ptr-match": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"s"},
		expect: "init",
	},
	"ptr-invalid": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"invalid"},
		expect: "default",
	},
	"ptr-any": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{},
		expect: "init",
	},
	"ptr-star": {
		param:  structPtrInit,
		deflt:  "default",
		names:  []string{"invalid", "*"},
		expect: "init",
	},

	// Test cases for panic scenarios.
	"panic-int-alias": {
		setup: test.Panic("cast failed [int]: 1"),
		call: func(_ findParams) any {
			return reflect.Find[any](intAlias(1), 0)
		},
	},
	"panic-ptr-alias": {
		setup: test.Panic(fmt.Sprintf(
			"cast failed [%T]: %v", structPtrAlias(nil), new(structAny))),
		param: new(structAny),
		call: func(_ findParams) any {
			return reflect.Find[*structAny, structPtrAlias](new(structAny), nil)
		},
	},
}

// TestFind tests the `Find` function of the reflect package using various test
// cases defined in the `findTestCases` map.
func TestFind(t *testing.T) {
	test.Map(t, findTestCases).
		Run(func(t test.Test, param findParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)
			if param.call == nil {
				param.call = findCall
			}

			// When
			expect := param.call(param)

			// Then
			assert.Equal(t, param.expect, expect)
		})
}

// nameParams is a test parameter type for testing the `Name` function of the
// reflect package.
type nameParams struct {
	name   string
	param  any
	expect string
}

// nameTestCases is a map of test parameters for testing the `Name` function of
// the reflect package.
var nameTestCases = map[string]nameParams{
	// Empty names.
	"empty-name-with-primitive": {
		name:  "",
		param: 42,
	},
	"empty-name-with-string": {
		name:   "",
		param:  "test",
		expect: "test",
	},

	// Provided names.
	"provided-name": {
		name:   "custom name",
		param:  struct{}{},
		expect: "custom-name",
	},
	"provided-name-with-spaces": {
		name:   "test case name",
		param:  struct{}{},
		expect: "test-case-name",
	},
	"provided-name-with-multiple-spaces": {
		name:   "test  case  name",
		param:  struct{}{},
		expect: "test--case--name",
	},
	"provided-name-already-hyphenated": {
		name:   "test-case-name",
		param:  struct{}{},
		expect: "test-case-name",
	},
	"provided-name-overrides-param": {
		name:   "override",
		param:  struct{ name string }{name: "ignored"},
		expect: "override",
	},

	// Parameter-based names.
	"empty-name-with-private-name-field": {
		name:   "",
		param:  struct{ name string }{name: "test case"},
		expect: "test-case",
	},
	"empty-name-with-exported-name-field": {
		name:   "",
		param:  struct{ Name string }{Name: "test case"},
		expect: "test-case",
	},
	"empty-name-without-name-field": {
		name:  "",
		param: struct{ value string }{value: "test"},
	},
	"empty-name-with-empty-name-field": {
		name:  "",
		param: struct{ name string }{name: ""},
	},
	"empty-name-with-nil-pointer": {
		name:  "",
		param: (*struct{ name string })(nil),
	},

	// Special parameter-based cases.
	"pointer-to-struct-with-name-field": {
		name:   "",
		param:  &struct{ name string }{name: "pointer test"},
		expect: "pointer-test",
	},
	"private-and-public-name-fields": {
		name: "",
		param: struct {
			name string
			Name string
		}{name: "private", Name: "public"},
		expect: "private",
	},
}

// TestName tests the `Name` function of the reflect package using various test
// cases defined in the `nameTestCases` map.
func TestName(t *testing.T) {
	test.Map(t, nameTestCases).
		Run(func(t test.Test, param nameParams) {
			// When
			result := reflect.Name(param.name, param.param)

			// Then
			assert.Equal(t, param.expect, result)
		})
}
