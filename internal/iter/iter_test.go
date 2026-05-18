package iter_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/iter"
	"github.com/tkrop/go-testing/test"
)

type SyncMapParams struct {
	input  map[string]int
	expect map[string]int
}

var syncMapTestCases = map[string]SyncMapParams{
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

func TestSyncMap(t *testing.T) {
	test.Map(t, syncMapTestCases).
		Run(func(t test.Test, param SyncMapParams) {
			// Given
			var source sync.Map
			for k, v := range param.input {
				source.Store(k, v)
			}

			// When
			result := map[string]int{}
			for k, v := range iter.SyncMap[string, int](&source) {
				result[k] = v
			}

			// Then
			assert.Equal(t, param.expect, result)
		})
}
