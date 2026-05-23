package maps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/maps"
	"github.com/tkrop/go-testing/test"
)

type CopyParams struct {
	target  map[string]int
	sources []map[string]int
	expect  map[string]int
}

var copyTestCases = map[string]CopyParams{
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

func TestCopy(t *testing.T) {
	test.Map(t, copyTestCases).
		Run(func(t test.Test, param CopyParams) {
			// When
			result := maps.Copy(param.target, param.sources...)

			// Then
			assert.Equal(t, param.expect, result)
		})
}
