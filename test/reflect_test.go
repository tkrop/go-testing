package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type Struct struct{ s string }

func NewStruct(s string) Struct     { return Struct{s: s} }
func NewPtrStruct(s string) *Struct { return &Struct{s: s} }

type testBuilderStructParam struct {
	target Struct
	setup  func(test.Builder[Struct])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[Struct])
}

var testBuilderStructParams = map[string]testBuilderStructParam{
	"test struct get is empty - no copy possible": {
		target: NewStruct("test get"),
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, NewStruct(""), b.Get(""))
			assert.Equal(t, NewStruct(""), b.Build())
		},
	},

	"test struct set": {
		target: NewStruct("test set"),
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewStruct("test set final"), b.Get(""))
			assert.Equal(t, NewStruct("test set final"), b.Build())
		},
	},

	"test struct reset no pointer": {
		target: NewStruct("test reset"),
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "test reset first").
				Set("", NewStruct("test reset final"))
		},
		expect: test.Panic("target must be a compatible struct pointer"),
	},

	"test struct reset pointer": {
		target: NewStruct("test reset"),
		setup: func(b test.Builder[Struct]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[Struct]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewStruct("test reset final"), b.Build())
		},
	},
}

func TestBuilderStruct(t *testing.T) {
	// test.New[testBuilderStructParam](t, testBuilderStructParams["test struct reset pointer"]).
	test.Map(t, testBuilderStructParams).
		Run(func(t test.Test, param testBuilderStructParam) {
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

type testBuilderPtrStructParam struct {
	target *Struct
	setup  func(test.Builder[*Struct])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[*Struct])
}

var testBuilderPtrStructParams = map[string]testBuilderPtrStructParam{
	"test nil any get": {
		target: nil,
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, NewPtrStruct(""), b.Get(""))
			assert.Equal(t, NewPtrStruct(""), b.Build())
		},
	},

	"test nil any set": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test set final"), b.Build())
		},
	},

	"test nil any reset": {
		target: nil,
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Build())
		},
	},

	"test nil struct get": {
		target: new(Struct),
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, NewPtrStruct(""), b.Get(""))
			assert.Equal(t, NewPtrStruct(""), b.Build())
		},
	},

	"test nil struct set": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test set final"), b.Build())
		},
	},

	"test nil struct reset": {
		target: new(Struct),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Build())
		},
	},

	"test ptr get": {
		target: NewPtrStruct("test get"),
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test get", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test get"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test get"), b.Build())
		},
	},

	"test ptr set": {
		target: NewPtrStruct("test set"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test set final"), b.Build())
		},
	},

	"test ptr reset": {
		target: NewPtrStruct("test reset"),
		setup: func(b test.Builder[*Struct]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[*Struct]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Build())
		},
	},
}

func TestBuilderPtrStruct(t *testing.T) {
	test.Map(t, testBuilderPtrStructParams).
		Run(func(t test.Test, param testBuilderPtrStructParam) {
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

type testBuilderAnyParam struct {
	target any
	setup  func(test.Builder[any])
	expect mock.SetupFunc
	check  func(test.Test, test.Builder[any])
}

var testBuilderAnyParams = map[string]testBuilderAnyParam{
	"test invalid type": {
		target: nil, // nil is type any.
		expect: test.Panic("target must be a struct or pointer to struct"),
	},

	"test struct get is empty - no copy possible": {
		target: NewStruct("test get"),
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, NewStruct(""), b.Get(""))
			assert.Equal(t, NewStruct(""), b.Build())
		},
	},

	"test struct set": {
		target: NewStruct("test set"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewStruct("test set final"), b.Get(""))
			assert.Equal(t, NewStruct("test set final"), b.Build())
		},
	},

	"test struct reset no pointer": {
		target: NewStruct("test reset"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test reset first").
				Set("", NewStruct("test reset final"))
		},
		expect: test.Panic("target must be a compatible struct pointer"),
	},

	"test struct reset pointer": {
		target: NewStruct("test reset"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewStruct("test reset final"), b.Build())
		},
	},

	"test ptr get": {
		target: NewPtrStruct("test get"),
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test get", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test get"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test get"), b.Build())
		},
	},

	"test ptr set": {
		target: NewPtrStruct("test set"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test set final"), b.Build())
		},
	},

	"test ptr reset": {
		target: NewPtrStruct("test reset"),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Build())
		},
	},

	"test nil get": {
		target: new(Struct),
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "", b.Get("s"))
			assert.Equal(t, NewPtrStruct(""), b.Get(""))
			assert.Equal(t, NewPtrStruct(""), b.Build())
		},
	},

	"test nil set": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test set final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test set final"), b.Build())
		},
	},

	"test nil reset": {
		target: new(Struct),
		setup: func(b test.Builder[any]) {
			b.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, b test.Builder[any]) {
			assert.Equal(t, "test reset final", b.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Get(""))
			assert.Equal(t, NewPtrStruct("test reset final"), b.Build())
		},
	},
}

func TestBuilderAny(t *testing.T) {
	test.Map(t, testBuilderAnyParams).
		Run(func(t test.Test, param testBuilderAnyParam) {
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

func TestNewBuilderStruct(t *testing.T) {
	// Given
	builder := test.NewBuilder[Struct]()

	// When
	builder.Set("s", "test set")

	// Then
	assert.Equal(t, "test set", builder.Get("s"))
	assert.Equal(t, NewStruct("test set"), builder.Get(""))
	assert.Equal(t, NewStruct("test set"), builder.Build())
}

func TestNewBuilderPtrStruct(t *testing.T) {
	// Given
	builder := test.NewBuilder[*Struct]()

	// When
	builder.Set("s", "test set")

	// Then
	assert.Equal(t, "test set", builder.Get("s"))
	assert.Equal(t, NewPtrStruct("test set"), builder.Get(""))
	assert.Equal(t, NewPtrStruct("test set"), builder.Build())
}
