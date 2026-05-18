package maps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/test"
)

type CopyParams struct {
	input  map[string]int
	expect map[string]int
}

var copyTestCases = map[string]CopyParams{
	"empty-map": {
		input:  map[string]int{},
		expect: map[string]int{},
	},
	"single-key-value-pair": {
		input:  map[string]int{"a": 1},
		expect: map[string]int{"a": 1},
	},
	"multiple-key-value-pairs": {
		input:  map[string]int{"a": 1, "b": 2, "c": 3},
		expect: map[string]int{"a": 1, "b": 2, "c": 3},
	},
}

func TestCopy(t *testing.T) {
	test.Map(t, copyTestCases).
		Run(func(t test.Test, param CopyParams) {
			// When
			expect := maps.Copy(param.input)

			// Then
			assert.Equal(t, param.expect, expect)
		})
}

type AddParams struct {
	target  map[string]int
	sources []map[string]int
	expect  map[string]int
}

var addTestCases = map[string]AddParams{
	"no-sources": {
		target:  map[string]int{"a": 1},
		sources: []map[string]int{},
		expect:  map[string]int{"a": 1},
	},
	"single-source": {
		target:  map[string]int{"a": 1},
		sources: []map[string]int{{"b": 2}},
		expect:  map[string]int{"a": 1, "b": 2},
	},
	"multiple-sources-with-no-conflicts": {
		target:  map[string]int{"a": 1},
		sources: []map[string]int{{"b": 2}, {"c": 3}},
		expect:  map[string]int{"a": 1, "b": 2, "c": 3},
	},
	"multiple-sources-with-conflicts": {
		target:  map[string]int{"a": 1},
		sources: []map[string]int{{"a": 2}, {"a": 3, "b": 4}},
		expect:  map[string]int{"a": 3, "b": 4},
	},
}

func TestAdd(t *testing.T) {
	test.Map(t, addTestCases).
		Run(func(t test.Test, param AddParams) {
			// When
			result := maps.Add(param.target, param.sources...)

			// Then
			assert.Equal(t, param.expect, result)
		})
}

type CollectParams struct {
	input  map[string]int
	expect map[string]int
}

var collectTestCases = map[string]CollectParams{
	"empty-map": {
		input:  map[string]int{},
		expect: map[string]int{},
	},
	"single-key-value-pair": {
		input:  map[string]int{"a": 1},
		expect: map[string]int{"a": 1},
	},
	"multiple-key-value-pairs": {
		input:  map[string]int{"a": 1, "b": 2, "c": 3},
		expect: map[string]int{"a": 1, "b": 2, "c": 3},
	},
}

func TestCollect(t *testing.T) {
	test.Map(t, collectTestCases).
		Run(func(t test.Test, param CollectParams) {
			// When
			result := maps.Collect(func(yield func(string, int) bool) {
				for k, v := range param.input {
					if !yield(k, v) {
						return
					}
				}
			})

			// Then
			assert.Equal(t, param.expect, result)
		})
}
