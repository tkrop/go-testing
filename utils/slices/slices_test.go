package slices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/testing/utils/slices"
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
