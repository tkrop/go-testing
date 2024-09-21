package test_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrop/go-testing/test"
)

type testErrorParam struct {
	error error
	setup func(*test.Accessor[error])
	test  func(test.Test, *test.Accessor[error])
}

var testErrorParams = map[string]testErrorParam{
	"test get": {
		error: errors.New("test get"),
		test: func(t test.Test, a *test.Accessor[error]) {
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
		test: func(t test.Test, a *test.Accessor[error]) {
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
			param.test(t, accessor)
		})
}
