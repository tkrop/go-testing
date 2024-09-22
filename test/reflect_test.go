package test_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type Struct struct{ s string }

func NewStruct(s string) Struct     { return Struct{s: s} }
func NewPtrStruct(s string) *Struct { return &Struct{s: s} }

type testAccessorParam struct {
	target any
	setup  func(*test.Accessor[any])
	expect mock.SetupFunc
	check  func(test.Test, *test.Accessor[any])
}

var testAccessorParams = map[string]testAccessorParam{
	"test invalid type": {
		target: int(1),
		expect: test.Panic("target must be a struct or pointer to struct"),
	},

	"test struct get is empty - no copy possible": {
		target: NewStruct("test get"),
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "", a.Get("s"))
			assert.Equal(t, NewStruct(""), a.Get(""))
		},
	},

	"test struct set": {
		target: NewStruct("test set"),
		setup: func(a *test.Accessor[any]) {
			a.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "test set final", a.Get("s"))
			assert.Equal(t, NewStruct("test set final"), a.Get(""))
		},
	},

	"test struct reset no pointer": {
		target: NewStruct("test reset"),
		setup: func(a *test.Accessor[any]) {
			a.Set("s", "test reset first").
				Set("", NewStruct("test reset final"))
		},
		expect: test.Panic("target must of compatible struct pointer type"),
	},

	"test struct reset pointer": {
		target: NewStruct("test reset"),
		setup: func(a *test.Accessor[any]) {
			a.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "test reset final", a.Get("s"))
			assert.Equal(t, NewStruct("test reset final"), a.Get(""))
		},
	},

	"test ptr get": {
		target: NewPtrStruct("test get"),
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "test get", a.Get("s"))
			assert.Equal(t, NewPtrStruct("test get"), a.Get(""))
		},
	},

	"test ptr set": {
		target: NewPtrStruct("test set"),
		setup: func(a *test.Accessor[any]) {
			a.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "test set final", a.Get("s"))
			assert.Equal(t, NewPtrStruct("test set final"), a.Get(""))
		},
	},

	"test ptr reset": {
		target: NewPtrStruct("test reset"),
		setup: func(a *test.Accessor[any]) {
			a.Set("s", "test reset first").
				Set("", NewPtrStruct("test reset final"))
		},
		check: func(t test.Test, a *test.Accessor[any]) {
			assert.Equal(t, "test reset final", a.Get("s"))
			assert.Equal(t, NewPtrStruct("test reset final"), a.Get(""))
		},
	},
}

func TestAccessor(t *testing.T) {
	test.Map(t, testAccessorParams).
		Run(func(t test.Test, param testAccessorParam) {
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

type testErrorParam struct {
	error error
	setup func(*test.Accessor[error])
	check func(test.Test, *test.Accessor[error])
}

var testErrorParams = map[string]testErrorParam{
	"test get": {
		error: errors.New("test get"),
		check: func(t test.Test, a *test.Accessor[error]) {
			assert.Equal(t, "test get", a.Get("s"))
			assert.Equal(t, errors.New("test get"), a.Get(""))
		},
	},

	"test set": {
		error: errors.New("test set"),
		setup: func(a *test.Accessor[error]) {
			a.Set("s", "test set first").
				Set("s", "test set final")
		},
		check: func(t test.Test, a *test.Accessor[error]) {
			assert.Equal(t, "test set final", a.Get("s"))
			assert.Equal(t, errors.New("test set final"), a.Get(""))
		},
	},

	"test reset": {
		error: errors.New("test set"),
		setup: func(a *test.Accessor[error]) {
			a.Set("s", "test set first").
				Set("", errors.New("test set final"))
		},
		check: func(t test.Test, a *test.Accessor[error]) {
			assert.Equal(t, "test set final", a.Get("s"))
			assert.Equal(t, errors.New("test set final"), a.Get(""))
		},
	},
}

func TestError(t *testing.T) {
	test.Map(t, testErrorParams).
		Run(func(t test.Test, param testErrorParam) {
			// Given
			accessor := test.Error(param.error)

			// When
			if param.setup != nil {
				param.setup(accessor)
			}

			// The
			param.check(t, accessor)
		})
}
