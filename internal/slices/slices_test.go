package slices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/slices"
	"github.com/tkrop/go-testing/test"
)

func TestReverse(t *testing.T) {
	// When
	result := slices.Reverse([]int{0, 1, 2, 3, 4})

	// Then
	assert.Equal(t, []int{4, 3, 2, 1, 0}, result)
}

func TestPermut(t *testing.T) {
	// When
	result := slices.Permute([]int{0, 1, 2})

	// Then
	assert.Equal(t, [][]int{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 1, 0},
		{2, 0, 1},
	}, result)
}

type TestAddIntParam struct {
	slices [][]int
	expect []int
}

var addIntTestCases = map[string]TestAddIntParam{
	"add-multiple-int-slices": {
		slices: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
		expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	},
	"add-two-int-slices": {
		slices: [][]int{{1, 2}, {3, 4}},
		expect: []int{1, 2, 3, 4},
	},
	"add-single-int-slice": {
		slices: [][]int{{1, 2, 3}},
		expect: []int{1, 2, 3},
	},
	"add-empty-int-slices": {
		slices: [][]int{{}, {1, 2}, {}},
		expect: []int{1, 2},
	},
	"add-all-empty-int-slices": {
		slices: [][]int{{}, {}, {}},
		expect: []int{},
	},
	"add-no-int-slices": {
		slices: [][]int{},
		expect: []int{},
	},
}

func TestAddInt(t *testing.T) {
	test.Map(t, addIntTestCases).
		Run(func(t test.Test, param TestAddIntParam) {
			// When
			result := slices.Add[int](param.slices...)
			// Then
			assert.Equal(t, param.expect, result)
		})
}

type TestAddStringParam struct {
	slices [][]string
	expect []string
}

var addStringTestCases = map[string]TestAddStringParam{
	"add-string-slices": {
		slices: [][]string{{"hello", "world"}, {"foo", "bar"}},
		expect: []string{"hello", "world", "foo", "bar"},
	},
	"add-empty-string-slices": {
		slices: [][]string{{}, {"test"}, {}},
		expect: []string{"test"},
	},
	"add-no-string-slices": {
		slices: [][]string{},
		expect: []string{},
	},
}

func TestAdd(t *testing.T) {
	test.Map(t, addStringTestCases).
		Run(func(t test.Test, param TestAddStringParam) {
			// When
			result := slices.Add(param.slices...)
			// Then
			assert.Equal(t, param.expect, result)
		})
}
