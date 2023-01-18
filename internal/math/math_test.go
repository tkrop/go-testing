package math_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/math"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

type MinMaxParams struct {
	setup  mock.SetupFunc
	values []int
	result int
	expect test.Expect
}

var testMinParams = map[string]MinMaxParams{
	"no-values": {
		setup: test.Panic("runtime error: index out of range [0] with length 0"),
	},

	"first-value": {
		values: []int{-2, -1, 0, 1, 2},
		result: -2,
		expect: test.Success,
	},
	"last-value": {
		values: []int{2, 1, 0, -1, -2},
		result: -2,
		expect: test.Success,
	},
	"middle-value": {
		values: []int{-1, 1, -2, 2, 0},
		result: -2,
		expect: test.Success,
	},
}

func TestMin(t *testing.T) {
	test.Map(t, testMinParams).
		Run(func(t test.Test, param MinMaxParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			result := math.Min(param.values...)

			// Then
			assert.Equal(t, param.result, result)
		})
}

var testMaxParams = map[string]MinMaxParams{
	"no-values": {
		setup: test.Panic("runtime error: index out of range [0] with length 0"),
	},

	"first-value": {
		values: []int{2, 1, 0, -1, -2},
		result: 2,
		expect: test.Success,
	},
	"last-value": {
		values: []int{-2, -1, 0, 1, 2},
		result: 2,
		expect: test.Success,
	},
	"middle-value": {
		values: []int{-1, 1, -2, 2, 0},
		result: 2,
		expect: test.Success,
	},
}

func TestMax(t *testing.T) {
	test.Map(t, testMaxParams).
		Run(func(t test.Test, param MinMaxParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			result := math.Max(param.values...)

			// Then
			assert.Equal(t, param.result, result)
		})
}
