package test_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/test"
)

type ErrorMatcherParams struct {
	base   any
	match  any
	result bool
}

var testErrorMatcherParams = map[string]ErrorMatcherParams{
	"success-string-string": {
		base:   "error",
		match:  "error",
		result: true,
	},
	"success-string-error": {
		base:   "error",
		match:  errors.New("error"),
		result: true,
	},
	"success-error-string": {
		base:   errors.New("error"),
		match:  "error",
		result: true,
	},
	"success-error-error": {
		base:   errors.New("error"),
		match:  errors.New("error"),
		result: true,
	},
	"success-other-other": {
		base:   1,
		match:  1,
		result: true,
	},

	"failure-string-string": {
		base:   "error",
		match:  "error-other",
		result: false,
	},
	"failure-string-error": {
		base:   "error",
		match:  errors.New("error-other"),
		result: false,
	},
	"failure-error-string": {
		base:   errors.New("error"),
		match:  "error-other",
		result: false,
	},
	"failure-error-error": {
		base:   errors.New("error"),
		match:  errors.New("error-other"),
		result: false,
	},
	"failure-other-other": {
		base:   1,
		match:  false,
		result: false,
	},
}

func TestErrorMatcher(t *testing.T) {
	test.Map(t, testErrorMatcherParams).
		Run(func(t test.Test, param ErrorMatcherParams) {
			// Given
			matcher := test.Error(param.base)

			// When
			result := matcher.Matches(param.match)

			// Then
			assert.Equal(t, param.result, result)
		})
}

func TestErrorMatcherString(t *testing.T) {
	assert.Equal(t, test.Error(true).String(), "is equal to true (bool)")
}
